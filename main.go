package main

import (
	"bs-core/protocol"
	_ "bs-core/protocol/client"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	reset   = "\033[0m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	gray    = "\033[90m"
	bold    = "\033[1m"
)

var sessionID int32
var mu sync.Mutex

func logPacket(dir string, packetID int32, size int) {
	arrow := cyan + ">>>" + reset
	if dir == "IN" {
		arrow = yellow + "<<<" + reset
	}
	t := time.Now().Format("15:04:05")
	name := "Unknown"
	if packetID == 10100 {
		name = "ClientHello"
	}
	fmt.Printf("%s%s %-5s %s %s%s%s (%d bytes)\n",
		gray, t,
		yellow+dir+reset,
		arrow,
		bold, name, reset,
		size)
}

func handleConn(conn net.Conn, sid int32) {
	conn.(*net.TCPConn).SetNoDelay(true)
	addr := conn.RemoteAddr().String()
	ip := strings.Split(addr, ":")[0]

	fmt.Printf("%s[S#%d]%s %s connected\n",
		magenta, sid, reset,
		cyan+ip+reset)

	ctx := &protocol.ClientContext{
		Conn:      conn,
		SessionID: sid,
		StartTime: time.Now(),
	}

	defer func() {
		dur := time.Since(ctx.StartTime).Round(time.Millisecond)
		fmt.Printf("%s[S#%d]%s %s disconnected %s%s%s\n",
			magenta, sid, reset,
			cyan+ip+reset,
			gray, dur, reset)
		conn.Close()
	}()

	buf := make([]byte, 4096)
	for {
		conn.SetReadDeadline(time.Now().Add(90 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		if n < 7 {
			continue
		}

		packetID := int32(buf[0])<<8 | int32(buf[1])
		length := int32(buf[2])<<16 | int32(buf[3])<<8 | int32(buf[4])

		if n < 7+int(length) {
			continue
		}

		payload := buf[7 : 7+length]
		logPacket("IN", packetID, n)

		if handler, ok := protocol.Handle(packetID); ok {
			handler(payload, ctx)
		}
	}
}

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("\nBye.")
		os.Exit(0)
	}()

	ln, err := net.Listen("tcp", ":9339")
	if err != nil {
		fmt.Printf("Fatal: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n%sListening on port 9339%s\n\n", green+bold, reset)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		mu.Lock()
		sessionID++
		sid := sessionID
		mu.Unlock()
		go handleConn(conn, sid)
	}
}
