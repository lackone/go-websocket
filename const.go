package go_websocket

import "time"

const (
	ReadLimit         = 1024
	ReadDeadline      = 10 * time.Second
	HeartbeatInterval = 5 * time.Second
	WriteDeadline     = 10 * time.Second
	PingMessage       = ""
	ReadBufferSize    = 1024
	WriteBufferSize   = 1024
)
