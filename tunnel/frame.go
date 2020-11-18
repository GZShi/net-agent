package tunnel

import (
	"encoding/binary"
	"io"
)

// Frame 数据传输的最小单位
type Frame struct {
	ID        uint32
	Type      uint8
	SessionID uint32
	Header    []byte // Header 一定是JSON结构
	DataType  uint8
	Data      []byte
}

// ID | Type | SessionID | DataType | len(Header) | len(Data) | Header | Data
const frameStableBufSize = 4 + 1 + 4 + 1 + 4 + 4

// Frame.Type 字典
const (
	// FrameStreamData 流式数据传输
	FrameStreamData = uint8(iota)
	// FrameRequest 请求帧
	FrameRequest
	// FrameResponse 应答帧
	FrameResponse
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

	return readed, nil
}
