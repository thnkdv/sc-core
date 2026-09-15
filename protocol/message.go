package protocol

import (
	"bs-core/datastream"
	"net"
)

type Message struct {
	*titan.ByteStream
	ID      int32
	Version int32
}

func New(data []byte) *Message {
	return &Message{
		ByteStream: titan.New(data),
	}
}

func (m *Message) Encode() {}

func Send(msg *Message, conn net.Conn) {
	if msg.ID < 20000 {
		return
	}
	msg.Encode()

	payload := msg.GetBuffer()
	header := make([]byte, 7)
	header[0] = byte(msg.ID >> 8)
	header[1] = byte(msg.ID)
	header[2] = byte(len(payload) >> 16)
	header[3] = byte(len(payload) >> 8)
	header[4] = byte(len(payload))
	header[5] = byte(msg.Version >> 8)
	header[6] = byte(msg.Version)

	trailer := []byte{0xFF, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00}
	packet := make([]byte, 0, 7+len(payload)+7)
	packet = append(packet, header...)
	packet = append(packet, payload...)
	packet = append(packet, trailer...)

	conn.Write(packet)
}
