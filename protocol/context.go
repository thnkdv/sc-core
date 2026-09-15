package protocol

import (
	"net"
	"time"
)

type ClientContext struct {
	Conn      net.Conn
	SessionID int32
	StartTime time.Time
}
