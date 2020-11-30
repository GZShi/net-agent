package tunnel

import (
	"encoding/binary"
	"encoding/json"
	"io"
)

func (t *tunnel) NewFrame(typ uint8) *Frame {
	return &Frame{ID: t.NewID(), Type: typ, DataType: BinaryData}
}

// Frame 数据传输的最小单位
type Frame struct {
	ID        uint32
	Type      uint8
	SessionID uint32
	Header    []byte // Header 一定是JSON结构
	DataType  uint8
	Data      []byte
}

// Headers 头部信息
type Headers map[string]string

// ID | Type | SessionID | DataType | len(Header) | len(Data) | Header | Data
const frameStableBufSize = 4 + 1 + 4 + 1 + 4 + 4

// Frame.Type 字典
const (
	// FrameStreamData 流式数据传输
	FrameStreamData = uint8(iota)
	// FrameRequest 请求帧
	FrameRequest
	// FrameResponse 应答帧
	FrameResponseOK
	FrameResponseErr

	// FrameDialRequest 创建连接的请求
	FrameDialRequest
	FrameDialResponse
)

// Frame.DataType 字典
const (
	// BinaryData 二进制数据
	BinaryData = uint8(iota)
	// TextData 文本类型数据
	TextData
	// JSONData JSON类型数据
	JSONData
)

// WriteTo 将结构体序列化写入writer中
func (f *Frame) WriteTo(w io.Writer) (int64, error) {
	headerSize := 0
	if f.Header != nil {
		headerSize = len(f.Header)
	}
	dataSize := 0
	if f.Data != nil {
		dataSize = len(f.Data)
	}

	buf := make([]byte, frameStableBufSize)
	binary.BigEndian.PutUint32(buf[0:4], f.ID)
	buf[4] = f.Type
	binary.BigEndian.PutUint32(buf[5:9], f.SessionID)
	buf[9] = f.DataType
	binary.BigEndian.PutUint32(buf[10:14], uint32(headerSize))
	binary.BigEndian.PutUint32(buf[14:18], uint32(dataSize))

	var err error
	var wn int
	written := int64(0)

	wn, err = w.Write(buf)
	written += int64(wn)
	if err != nil {
		return written, err
	}

	wn, err = w.Write(f.Header)
	written += int64(wn)
	if err != nil {
		return written, err
	}

	wn, err = w.Write(f.Data)
	written += int64(wn)
	if err != nil {
		return written, err
	}

	// log.Get().Debug("written: ", f)

	return written, nil
}

// ReadFrom 从reader中读取一个frame的数据
func (f *Frame) ReadFrom(r io.Reader) (int64, error) {
	readed := int64(0)

	buf := make([]byte, frameStableBufSize)
	rn, err := io.ReadFull(r, buf)
	readed += int64(rn)
	if err != nil {
		return readed, err
	}

	f.ID = binary.BigEndian.Uint32(buf[0:4])
	f.Type = buf[4]
	f.SessionID = binary.BigEndian.Uint32(buf[5:9])
	f.DataType = buf[9]
	headerSize := int(binary.BigEndian.Uint32(buf[10:14]))
	dataSize := int(binary.BigEndian.Uint32(buf[14:18]))

	varSize := headerSize + dataSize
	if varSize > 0 {
		varBuf := make([]byte, varSize)
		rn, err = io.ReadFull(r, varBuf)
		readed += int64(rn)
		if err != nil {
			return readed, err
		}

		if headerSize > 0 {
			f.Header = varBuf[0:headerSize]
		}
		if dataSize > 0 {
			f.Data = varBuf[headerSize:]
		}
	}

	// log.Get().Debug("readed: ", f)

	return readed, nil
}

// WriteHeader 将字典写入header
func (f *Frame) WriteHeader(m map[string]string) error {
	buf, err := json.Marshal(m)
	if err != nil {
		return err
	}
	f.Header = buf
	return nil
}

// ReadHeader 解析header
func (f *Frame) ReadHeader() (map[string]string, error) {
	m := make(map[string]string)
	if f.Header == nil {
		return m, nil
	}
	err := json.Unmarshal(f.Header, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
