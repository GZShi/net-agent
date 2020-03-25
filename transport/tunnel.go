package transport

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/GZShi/net-agent/logger"
	"github.com/sirupsen/logrus"
)

type cid uint32

func makeConnIDGenerator() func() cid {
	bigRound := int64(7 * 24 * 60 * 60)
	a := uint32(time.Now().Unix() % bigRound)
	count := a << 10

	return func() cid {
		return cid(atomic.AddUint32(&count, 1))
	}
}

// Tunnel 用于传递数据的隧道
type Tunnel struct {
	id          uint32
	logDetail   bool
	conn        net.Conn
	writeLocker sync.Mutex
	name        string
	secretName  string // 打码的名字，用于管理端展示，防止不同的隧道之间恶意连接
	randKey     []byte
	encStream   cipher.Stream // 用这个流来加密数据，采用异或xor方法，不改变数据长度
	decStream   cipher.Stream // 用这个流来解密数据
	newConnID   func() cid

	ports             sync.Map
	activePortCount   int32
	finishedPortCount int32
	failedPortCount   int32
	donePorts         []*dataPort
	donePortsMut      sync.Mutex
	lastHeartbeat     uint32

	created       time.Time
	finished      time.Time
	heartbeatTime time.Time
	uploadSize    uint64
	downloadSize  uint64
	uploadPack    uint64
	downloadPack  uint64

	currTransData *Data
	currTransTime time.Time
}

var tunnelIDSequence = uint32(0)

// NewTunnel 构造数据隧道
func NewTunnel(conn net.Conn, name, secret string, randKey []byte, logDetail bool) (*Tunnel, error) {
	if len(name) <= 6 {
		return nil, errors.New("tunnal.name must large than 6")
	}
	secretName := name[0:2] + "***" + name[len(name)-2:]

	blockCipher, err := aes.NewCipher(randKey)
	if err != nil {
		return nil, err
	}
	iv := []byte(secret)
	for len(iv) < blockCipher.BlockSize() {
		iv = append(iv, '*')
	}

	// 两个流是一样的，因为是读写同时进行，所以需要两个流，否则在数据链比较大时，会出现加解密混乱的bug
	encStream := cipher.NewCTR(blockCipher, iv[0:blockCipher.BlockSize()])
	decStream := cipher.NewCTR(blockCipher, iv[0:blockCipher.BlockSize()])

	return &Tunnel{
		id:                atomic.AddUint32(&tunnelIDSequence, 1),
		logDetail:         logDetail,
		conn:              conn,
		newConnID:         makeConnIDGenerator(),
		name:              name,
		secretName:        secretName,
		randKey:           randKey,
		encStream:         encStream,
		decStream:         decStream,
		activePortCount:   0,
		finishedPortCount: 0,
		failedPortCount:   0,
		lastHeartbeat:     0,

		created:       time.Now(),
		heartbeatTime: time.Now(),
		uploadSize:    0,
		downloadSize:  0,
		uploadPack:    0,
		downloadPack:  0,
	}, nil
}

// GetName 获取通道名字
func (t *Tunnel) GetName() string {
	return t.name
}

// Close 关闭通道
func (t *Tunnel) Close() {
	t.conn.Close()
}

// EncWrite 加密写
func (t *Tunnel) EncWrite(buf []byte) (int, error) {
	t.encStream.XORKeyStream(buf, buf)
	return t.conn.Write(buf)
}

// DecReadFull 解密读
func (t *Tunnel) DecReadFull(buf []byte) (int, error) {
	rn, err := io.ReadFull(t.conn, buf)
	if rn > 0 {
		t.decStream.XORKeyStream(buf, buf)
	}
	return rn, err
}

// Serve 开启传输服务
func (t *Tunnel) Serve() {
	// 不断 Read 和 Write，直到连接关闭
	defer t.Close()

	for {
		var data Data
		if err := t.readData(&data); err != nil {
			log.Get().WithError(err).Error("read data failed")
			break
		}

		switch data.Cmd {
		case cmdDialReq:
			go t.dialByData(&data)
		case cmdDialSuccess, cmdDialFailed:
			go t.dialCallbackWithData(&data)
		case cmdClose:
			go t.closeByData(&data)
		case cmdData:
			t.transportData(&data)
		case cmdHeartbeat:
			go func() {
				t.lastHeartbeat = uint32(data.ConnID)
				t.heartbeatTime = time.Now()
				// 收到心跳包请求，15秒后回应一个心跳包
				<-time.After(time.Second * 15)
				t.writeData(newHeartbeatData(data.ConnID + 1))
			}()
		case cmdTextMessages:
			go func() {
				// 收到来自对端的信息
				log.Get().WithField("text", string(data.Bytes)).Info("text message from peer")
			}()
		}
	}
}

// StartHeartbeat 开始心跳
func (t *Tunnel) StartHeartbeat() {
	t.writeData(newHeartbeatData(cid(0)))
}

// writeData 网通道里面写入数据包，binary版本
func (t *Tunnel) writeData(data Data) (total, writedBytes int, err error) {
	t.writeLocker.Lock()
	defer t.writeLocker.Unlock()
	defer func() {
		t.uploadSize += uint64(total)
	}()

	// cid(4) + cmd(1) + datalen(2:65535)
	header := make([]byte, 4+1+2)
	binary.BigEndian.PutUint32(header[0:4], uint32(data.ConnID))
	header[4] = data.Cmd

	datalen := 0
	if data.Bytes != nil {
		datalen = len(data.Bytes)
	}
	binary.BigEndian.PutUint16(header[5:7], uint16(datalen))

	total = 0

	wn, err := t.conn.Write(header)
	if wn > 0 {
		total += wn
	}
	if err != nil {
		return
	}
	if datalen > 0 {
		// t.encStream.XORKeyStream(data.Bytes, data.Bytes)
		// wn, err = t.conn.Write(data.Bytes)
		wn, err = t.EncWrite(data.Bytes)
		if wn > 0 {
			total += wn
			writedBytes = wn
		}
		if err != nil {
			return
		}
	}

	t.uploadPack++
	err = nil
	return
}

func (t *Tunnel) readData(data *Data) error {
	header := make([]byte, 7)

	rn, err := io.ReadFull(t.conn, header)
	if rn > 0 {
		t.downloadSize += uint64(rn)
	}
	if err != nil {
		return err
	}

	data.ConnID = cid(binary.BigEndian.Uint32(header[0:4]))
	data.Cmd = header[4]

	datalen := binary.BigEndian.Uint16(header[5:7])
	if datalen > 0 {
		data.Bytes = make([]byte, datalen)
		rn, err := t.DecReadFull(data.Bytes)
		// rn, err := io.ReadFull(t.conn, data.Bytes)
		if rn > 0 {
			t.downloadSize += uint64(rn)
			// t.decStream.XORKeyStream(data.Bytes, data.Bytes)
		}
		if err != nil {
			return err
		}
	}

	t.downloadPack++
	return nil
}

// dialByData 根据读取到的dial数据包发起请求
func (t *Tunnel) dialByData(data *Data) {
	addr := string(data.Bytes)
	port := newDataPort(data.ConnID, "", addr, "")
	defer port.close()

	if err := port.dialToNet(t); err != nil {
		if t.logDetail {
			log.Get().WithFields(logrus.Fields{
				"to":   addr,
				"from": "",
			}).WithError(err).Error("dial to net failed")
		}
		t.writeData(newDialAns(data.ConnID, err))
		return
	}
	if t.logDetail {
		log.Get().WithFields(logrus.Fields{
			"to":   addr,
			"from": "",
		}).Info("dial to net success")
	}

	atomic.AddInt32(&t.activePortCount, 1)
	t.ports.Store(data.ConnID, port)
	defer func() {
		t.ports.Delete(data.ConnID)
		atomic.AddInt32(&t.activePortCount, -1)
	}()

	if _, _, err := t.writeData(newDialAns(data.ConnID, nil)); err != nil {
		return
	}

	// blocked
	port.serve()

	go t.writeData(newCloseData(data.ConnID))
}

func (t *Tunnel) closeByData(data *Data) {
	value, exist := t.ports.Load(data.ConnID)
	if !exist {
		return
	}

	port := value.(*dataPort)
	port.close()
}

func (t *Tunnel) transportData(data *Data) {
	t.currTransData = data
	t.currTransTime = time.Now()
	defer func() {
		t.currTransData = nil
	}()
	value, exist := t.ports.Load(data.ConnID)
	if !exist {
		go t.writeData(newCloseData(data.ConnID))
		return
	}
	port := value.(*dataPort)
	_, err := port.writeDataToPorter(data.Bytes)
	if err != nil {
		go t.writeData(newCloseData(data.ConnID))
		return
	}
}

func (t *Tunnel) dialCallbackWithData(data *Data) {
	value, exist := t.ports.Load(data.ConnID)
	if !exist {
		return
	}
	var err error
	switch data.Cmd {
	case cmdDialFailed:
		err = errors.New(string(data.Bytes))
	case cmdDialSuccess:
		err = nil
	default:
		return
	}
	port := value.(*dataPort)
	port.dialCallback(err)
}

// Dial 与net.Dial一致的创建连接方式
func (t *Tunnel) Dial(sourceAddr, network, addr, userName string) (target net.Conn, err error) {
	// wait dial result
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		id := t.newConnID()
		port := newDataPort(id, sourceAddr, addr, userName)
		defer port.close()

		t.ports.Store(id, port)
		atomic.AddInt32(&t.activePortCount, 1)
		defer func() {
			atomic.AddInt32(&t.activePortCount, -1)
			t.ports.Delete(port.connID)

			// move to donePorts
			t.donePortsMut.Lock()
			t.donePorts = append(t.donePorts, port)
			t.donePortsMut.Unlock()
		}()

		target, err = port.dialWithTunnel(t)
		wg.Done()
		if err != nil {
			atomic.AddInt32(&t.failedPortCount, 1)
			return
		}

		port.serve()
		atomic.AddInt32(&t.finishedPortCount, 1)
		go t.writeData(newCloseData(port.connID))
	}()

	wg.Wait()
	return target, err
}

// SendTextMessage 向对端发送一条消息，展示在对方控制台上
func (t *Tunnel) SendTextMessage(text string) {
	t.writeData(newTextMessageData(t.newConnID(), text))
}
