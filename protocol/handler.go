package protocol

type MessageHandler func(data []byte, ctx *ClientContext)

var handlers = map[int32]MessageHandler{}

func Register(id int32, h MessageHandler) {
	handlers[id] = h
}

func Handle(id int32) (MessageHandler, bool) {
	h, ok := handlers[id]
	return h, ok
}
