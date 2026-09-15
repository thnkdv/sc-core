package titan

import (
	"encoding/binary"
	"math"
)

type ByteStream struct {
	Buffer    []byte
	Offset    int
	BitOffset int
}

func New(data []byte) *ByteStream {
	if data == nil {
		data = []byte{}
	}
	return &ByteStream{Buffer: data}
}

func (bs *ByteStream) ReadInt() int32 {
	bs.BitOffset = 0
	if bs.Offset+4 > len(bs.Buffer) {
		return 0
	}
	val := int32(bs.Buffer[bs.Offset])<<24 |
		int32(bs.Buffer[bs.Offset+1])<<16 |
		int32(bs.Buffer[bs.Offset+2])<<8 |
		int32(bs.Buffer[bs.Offset+3])
	bs.Offset += 4
	return val
}

func (bs *ByteStream) ReadShort() int16 {
	bs.BitOffset = 0
	if bs.Offset+2 > len(bs.Buffer) {
		return 0
	}
	val := int16(bs.Buffer[bs.Offset])<<8 | int16(bs.Buffer[bs.Offset+1])
	bs.Offset += 2
	return val
}

func (bs *ByteStream) ReadByte() byte {
	if bs.Offset >= len(bs.Buffer) {
		return 0
	}
	b := bs.Buffer[bs.Offset]
	bs.Offset++
	return b
}

func (bs *ByteStream) ReadString() string {
	length := bs.ReadInt()
	if length > 0 && length < 90000 {
		if bs.Offset+int(length) > len(bs.Buffer) {
			return ""
		}
		s := string(bs.Buffer[bs.Offset : bs.Offset+int(length)])
		bs.Offset += int(length)
		return s
	}
	return ""
}

func (bs *ByteStream) ReadVInt() int32 {
	var result int32
	var shift uint

	for {
		byteVal := int32(bs.ReadByte())
		if shift == 0 {
			a1 := (byteVal & 0x40) >> 6
			a2 := (byteVal & 0x80) >> 7
			s := (byteVal << 1) & ^int32(0x181)
			byteVal = s | (a2 << 7) | a1
		}
		result |= (byteVal & 0x7f) << shift
		shift += 7
		if byteVal&0x80 == 0 {
			break
		}
	}
	return (result >> 1) ^ (-(result & 1))
}

func (bs *ByteStream) ReadDataReference() [2]int32 {
	a := bs.ReadVInt()
	if a == 0 {
		return [2]int32{0, 0}
	}
	return [2]int32{a, bs.ReadVInt()}
}

func (bs *ByteStream) ReadLogicLong() [2]int32 {
	return [2]int32{bs.ReadVInt(), bs.ReadVInt()}
}

func (bs *ByteStream) ReadBoolean() bool {
	return bs.ReadVInt() >= 1
}

func (bs *ByteStream) ReadBytes() []byte {
	length := bs.ReadInt()
	if length <= 0 || int(length) > len(bs.Buffer)-bs.Offset {
		return nil
	}
	data := bs.Buffer[bs.Offset : bs.Offset+int(length)]
	bs.Offset += int(length)
	return data
}

func (bs *ByteStream) WriteInt(value int32) {
	bs.BitOffset = 0
	bs.ensureCapacity(4)
	bs.Buffer[bs.Offset] = byte(value >> 24)
	bs.Buffer[bs.Offset+1] = byte(value >> 16)
	bs.Buffer[bs.Offset+2] = byte(value >> 8)
	bs.Buffer[bs.Offset+3] = byte(value)
	bs.Offset += 4
}

func (bs *ByteStream) WriteShort(value int16) {
	bs.BitOffset = 0
	bs.ensureCapacity(2)
	bs.Buffer[bs.Offset] = byte(value >> 8)
	bs.Buffer[bs.Offset+1] = byte(value)
	bs.Offset += 2
}

func (bs *ByteStream) WriteByte(value byte) {
	bs.BitOffset = 0
	bs.ensureCapacity(1)
	bs.Buffer[bs.Offset] = value
	bs.Offset++
}

func (bs *ByteStream) WriteString(value string) {
	if len(value) == 0 || len(value) > 90000 {
		bs.WriteInt(-1)
		return
	}
	data := []byte(value)
	bs.WriteInt(int32(len(data)))
	bs.ensureCapacity(len(data))
	copy(bs.Buffer[bs.Offset:], data)
	bs.Offset += len(data)
}

func (bs *ByteStream) WriteVInt(value int32) {
	bs.BitOffset = 0
	temp := byte((value >> 25) & 0x40)
	flipped := value ^ (value >> 31)
	temp |= byte(value & 0x3F)
	value >>= 6
	flipped >>= 6
	if flipped == 0 {
		bs.WriteByte(temp)
		return
	}
	bs.WriteByte(temp | 0x80)
	flipped >>= 7
	var r byte
	if flipped != 0 {
		r = 0x80
	}
	bs.WriteByte(byte(value&0x7F) | r)
	value >>= 7
	for flipped != 0 {
		flipped >>= 7
		r = 0
		if flipped != 0 {
			r = 0x80
		}
		bs.WriteByte(byte(value&0x7F) | r)
		value >>= 7
	}
}

func (bs *ByteStream) WriteBoolean(value bool) {
	if bs.BitOffset == 0 {
		bs.ensureCapacity(1)
		bs.Buffer[bs.Offset] = 0
		bs.Offset++
	}
	if value {
		bs.Buffer[bs.Offset-1] |= 1 << uint(bs.BitOffset)
	}
	bs.BitOffset = (bs.BitOffset + 1) & 7
}

func (bs *ByteStream) WriteLong(v1, v2 int32) {
	bs.WriteInt(v1)
	bs.WriteInt(v2)
}

func (bs *ByteStream) WriteLogicLong(v1, v2 int32) {
	bs.WriteVInt(v1)
	bs.WriteVInt(v2)
}

func (bs *ByteStream) WriteDataReference(v1, v2 int32) {
	if v1 < 1 {
		bs.WriteVInt(0)
	} else {
		bs.WriteVInt(v1)
		bs.WriteVInt(v2)
	}
}

func (bs *ByteStream) WriteBytes(data []byte) {
	if data != nil {
		bs.WriteInt(int32(len(data)))
		bs.ensureCapacity(len(data))
		copy(bs.Buffer[bs.Offset:], data)
		bs.Offset += len(data)
		return
	}
	bs.WriteInt(-1)
}

func (bs *ByteStream) GetBuffer() []byte {
	return bs.Buffer[:bs.Offset]
}

func (bs *ByteStream) Hex() string {
	buf := bs.Buffer[:bs.Offset]
	hex := make([]byte, len(buf)*2)
	for i, b := range buf {
		hex[i*2] = "0123456789abcdef"[b>>4]
		hex[i*2+1] = "0123456789abcdef"[b&0x0f]
	}
	return string(hex)
}

func (bs *ByteStream) ensureCapacity(n int) {
	for bs.Offset+n > len(bs.Buffer) {
		bs.Buffer = append(bs.Buffer, make([]byte, 64)...)
	}
}

func init() {
	_ = math.MaxInt32
	_ = binary.BigEndian
}
