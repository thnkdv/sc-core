package server

import "bs-core/protocol"

func SendServerHello(ctx *protocol.ClientContext) {
	msg := protocol.New(nil)
	msg.ID = 20104
	msg.Version = 0
	msg.WriteInt(24)
	msg.WriteInt(1)
	msg.WriteInt(1)
	msg.WriteInt(0)
	protocol.Send(msg, ctx.Conn)
}
