package wsDefine

import (
	"time"
)

const (
	KeepAliveInterval = 15 * time.Second //保持活动连接的间隔
)

type ReadPrivateHandler func(msg []byte)
type ReadMarketHandler func(len int, bufPtr *[]byte)
type PingFunc func(*SafeWrite) error

type ReConnType uint8

const (
	READ_ERROR ReConnType = iota
	START_CONN
)

func (s ReConnType) String() string {
	switch s {
	case READ_ERROR:
		return "READ_ERROR"
	case START_CONN:
		return "START_CONN"
	default:
		return "ERROR"
	}
}
