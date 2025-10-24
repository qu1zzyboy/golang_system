package wsDefine

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type SafeWrite struct {
	conn   *websocket.Conn // websocket连接
	connMu sync.Mutex      // 互斥锁,保护conn的写并发
}

func NewSafeWrite(conn *websocket.Conn) *SafeWrite {
	return &SafeWrite{
		conn:   conn,
		connMu: sync.Mutex{},
	}
}

func (s *SafeWrite) GetConn() *websocket.Conn {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	return s.conn
}

func (s *SafeWrite) SafeWriteMsg(messageType int, data []byte) error {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	return s.conn.WriteMessage(messageType, data)
}

func (s *SafeWrite) SafeClose() error {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

func (s *SafeWrite) SafeWriteControl(messageType int, data []byte, deadline time.Time) error {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	return s.conn.WriteControl(messageType, data, deadline)
}
