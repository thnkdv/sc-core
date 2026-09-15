package client

import (
	"bs-core/protocol"
	"bs-core/protocol/server"
	"bs-core/datastream"
)

func init() {
	protocol.Register(10100, Handle)
}

func Handle(data []byte, ctx *protocol.ClientContext) {
	_ = titan.New(data)
	server.SendServerHello(ctx)
}
