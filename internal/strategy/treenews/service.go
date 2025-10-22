package treenews

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hhh500/quantGoInfra/infra/observe/log/dynamicLog"
	"github.com/hhh500/quantGoInfra/infra/safex"
)

var (
	logger    = dynamicLog.Log
	loggerErr = dynamicLog.Error
)

var (
	svcOnce     sync.Once
	svcInstance *Service
)

// GetService returns the singleton instance used by the boot module.
func GetService() *Service {
	svcOnce.Do(func() {
		svcInstance = NewService(defaultConfig())
	})
	return svcInstance
}

// Event represents a filtered Tree News payload that targets the Upbit KRW flow.
type Event struct {
	ID          string
	Symbols     []string
	Payload     map[string]any
	ReceivedAt  time.Time
	ServerMilli int64
}

// HandlerFunc is invoked for every filtered event.
type HandlerFunc func(context.Context, Event)

var (
	handlerMu sync.RWMutex
	handlers  []HandlerFunc
)

// RegisterHandler registers a callback that receives filtered Tree News events.
func RegisterHandler(fn HandlerFunc) {
	handlerMu.Lock()
	defer handlerMu.Unlock()
	handlers = append(handlers, fn)
}

// Service manages websocket workers and dispatches events to handlers.
type Service struct {
	cfg     Config
	dialer  *websocket.Dialer
	started atomic.Bool

	cancel context.CancelFunc
	wg     sync.WaitGroup

	dedup *idSet
}

func NewService(cfg Config) *Service {
	return &Service{
		cfg:    cfg,
		dialer: &websocket.Dialer{Proxy: http.ProxyFromEnvironment, HandshakeTimeout: 15 * time.Second},
		dedup:  newIDSet(cfg.DedupCapacity),
	}
}

func (s *Service) Start(ctx context.Context) error {
	if !s.cfg.Enabled {
		logger.GetLog().Info("tree news service disabled")
		return nil
	}
	if s.cfg.APIKey == "" {
		return errors.New("tree news api key empty")
	}
	if s.started.Load() {
		return nil
	}
	s.started.Store(true)

	ctxRun, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	if ctx != nil {
		safex.SafeGo("treenews_cancel", func() {
			select {
			case <-ctx.Done():
				cancel()
			case <-ctxRun.Done():
			}
		})
	}

	workers := s.cfg.Workers
	if workers <= 0 {
		workers = 1
	}
	for i := 0; i < workers; i++ {
		idx := i
		s.wg.Add(1)
		safex.SafeGo(fmt.Sprintf("treenews_worker_%d", idx), func() {
			defer s.wg.Done()
			s.worker(ctxRun, idx)
		})
	}
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	if !s.started.Load() {
		return nil
	}
	if s.cancel != nil {
		s.cancel()
	}
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}
	s.started.Store(false)
	return nil
}

func (s *Service) worker(ctx context.Context, workerID int) {
	backoff := 300 * time.Millisecond
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn, err := s.dial(ctx)
		if err != nil {
			loggerErr.GetLog().Errorf("tree news dial failed: %v", err)
			backoff = minDuration(backoff*2, 30*time.Second)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return
			}
			continue
		}
		backoff = 300 * time.Millisecond

		if err := s.login(conn); err != nil {
			loggerErr.GetLog().Errorf("tree news login failed: %v", err)
			conn.Close()
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return
			}
			continue
		}

		errCh := make(chan error, 1)
		go s.readLoop(ctx, conn, errCh)
		go s.heartbeatLoop(ctx, conn, errCh)
		go s.rollingLoop(ctx, conn, errCh)

		select {
		case <-ctx.Done():
			conn.Close()
			return
		case err := <-errCh:
			if err != nil && !errors.Is(err, context.Canceled) {
				loggerErr.GetLog().Errorf("tree news worker error: %v", err)
			}
			conn.Close()
		}
	}
}

func (s *Service) dial(ctx context.Context) (*websocket.Conn, error) {
	conn, _, err := s.dialer.DialContext(ctx, s.cfg.URL, nil)
	return conn, err
}

func (s *Service) login(conn *websocket.Conn) error {
	message := fmt.Sprintf("login %s", s.cfg.APIKey)
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return conn.WriteMessage(websocket.TextMessage, []byte(message))
}

func (s *Service) readLoop(ctx context.Context, conn *websocket.Conn, errCh chan<- error) {
	conn.SetReadLimit(2 << 20)
	if s.cfg.PingTimeout > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(s.cfg.PingInterval + s.cfg.PingTimeout))
		conn.SetPongHandler(func(string) error {
			return conn.SetReadDeadline(time.Now().Add(s.cfg.PingInterval + s.cfg.PingTimeout))
		})
	}

	for {
		if ctx.Err() != nil {
			return
		}
		_, data, err := conn.ReadMessage()
		if err != nil {
			errCh <- err
			return
		}
		if s.cfg.PingTimeout > 0 {
			_ = conn.SetReadDeadline(time.Now().Add(s.cfg.PingInterval + s.cfg.PingTimeout))
		}
		if err := s.handleMessage(ctx, data); err != nil {
			loggerErr.GetLog().Errorf("tree news handle message failed: %v", err)
		}
	}
}

func (s *Service) heartbeatLoop(ctx context.Context, conn *websocket.Conn, errCh chan<- error) {
	if s.cfg.PingInterval <= 0 {
		return
	}
	ticker := time.NewTicker(s.cfg.PingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				errCh <- err
				return
			}
		}
	}
}

func (s *Service) rollingLoop(ctx context.Context, conn *websocket.Conn, errCh chan<- error) {
	if s.cfg.RollingReconnect <= 0 {
		return
	}
	jitter := time.Duration(rand.Int63n(int64(s.cfg.RollingJitter)))
	timer := time.NewTimer(s.cfg.RollingReconnect + jitter)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return
	case <-timer.C:
		errCh <- errors.New("rolling reconnect")
		return
	}
}

func (s *Service) handleMessage(ctx context.Context, data []byte) error {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}
	id := toString(payload["_id"])
	if id == "" {
		return nil
	}
	if !s.dedup.Add(id) {
		return nil
	}

	symbols := upbitKRWSymbols(payload)
	if len(symbols) == 0 {
		return nil
	}

	event := Event{
		ID:         id,
		Symbols:    symbols,
		Payload:    payload,
		ReceivedAt: time.Now().UTC(),
	}
	if ts := toInt64(payload["time"]); ts > 0 {
		event.ServerMilli = ts
	}

	handlerMu.RLock()
	defer handlerMu.RUnlock()
	for _, h := range handlers {
		h(ctx, event)
	}
	return nil
}

type idSet struct {
	mu    sync.Mutex
	order []string
	set   map[string]struct{}
	max   int
}

func newIDSet(capacity int) *idSet {
	if capacity <= 0 {
		capacity = 1000
	}
	return &idSet{
		order: make([]string, 0, capacity),
		set:   make(map[string]struct{}, capacity),
		max:   capacity,
	}
}

func (s *idSet) Add(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.set[id]; ok {
		return false
	}
	s.set[id] = struct{}{}
	s.order = append(s.order, id)
	if len(s.order) > s.max {
		stale := s.order[0]
		s.order = s.order[1:]
		delete(s.set, stale)
	}
	return true
}

func toString(v any) string {
	switch vv := v.(type) {
	case string:
		return vv
	case fmt.Stringer:
		return vv.String()
	case json.Number:
		return vv.String()
	case float64:
		return fmt.Sprintf("%.0f", vv)
	case int64:
		return fmt.Sprintf("%d", vv)
	case int:
		return fmt.Sprintf("%d", vv)
	default:
		return ""
	}
}

func toInt64(v any) int64 {
	switch vv := v.(type) {
	case int64:
		return vv
	case int:
		return int64(vv)
	case float64:
		return int64(vv)
	case json.Number:
		if i, err := vv.Int64(); err == nil {
			return i
		}
	case string:
		if i, err := strconv.ParseInt(vv, 10, 64); err == nil {
			return i
		}
	}
	return 0
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
