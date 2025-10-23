// Package treenews 负责监听 Tree News WebSocket，筛选 Upbit KRW 相关事件，并以 channel 模式推送给业务层。
package treenews

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hhh500/quantGoInfra/infra/observe/log/dynamicLog"
	"github.com/hhh500/quantGoInfra/infra/observe/log/staticLog"
	"github.com/hhh500/quantGoInfra/infra/safex"
	"github.com/hhh500/quantGoInfra/pkg/utils/timeUtils"
)

var (
	treeLogConfig = staticLog.Config{
		NeedErrorHook: true,
		FileDir:       "tree_news_log",
		DateStr:       timeUtils.GetNowDateStr(),
		FileName:      "tree_news",
		Level:         staticLog.INFO_LEVEL,
	}
	treeErrConfig = staticLog.Config{
		NeedErrorHook: false,
		FileDir:       "tree_news_log",
		DateStr:       timeUtils.GetNowDateStr(),
		FileName:      "tree_news_error",
		Level:         staticLog.ERROR_LEVEL,
	}
	logger    = dynamicLog.NewDynamicLogger(treeLogConfig)
	loggerErr = dynamicLog.NewDynamicLogger(treeErrConfig)
)

var (
	svcOnce     sync.Once
	svcInstance *Service
)

// GetService 返回供 boot 模块使用的单例。
func GetService() *Service {
	svcOnce.Do(func() {
		svcInstance = NewService(defaultConfig())
	})
	return svcInstance
}

// Event 表示通过筛选的 Tree News 事件，聚焦 Upbit KRW 场景。
type Event struct {
	ID              string
	Symbols         []string
	Payload         map[string]any
	ReceivedAt      time.Time
	ServerMilli     int64
	LatencyRawMS    int
	LatencyAdjustMS int
	LatencyMS       int
	RTTMS           int64
}

// HandlerFunc 会在每条过滤后的事件上被调用。
type HandlerFunc func(context.Context, Event)

var (
	handlerMu sync.RWMutex
	handlers  []HandlerFunc
)

// RegisterHandler 注册事件回调。
func RegisterHandler(fn HandlerFunc) {
	handlerMu.Lock()
	defer handlerMu.Unlock()
	handlers = append(handlers, fn)
}

// Service 管理 WebSocket 工作者并将事件分发给回调。
// Service 维护 WebSocket 连接、读写协程和消息去重等核心状态。
type Service struct {
	cfg     Config
	dialer  *websocket.Dialer
	started atomic.Bool

	cancel context.CancelFunc // cancel 用于通知所有后台协程退出
	wg     sync.WaitGroup     // wg 等待后台协程全部结束

	dedup            *idSet // dedup 保存最近的消息 id，防止重复处理
	outQueue         chan queuedMessage
	latencyWarnMS    int
	latencyWarnCount int
	rttWarnMS        int
	rttWarnCount     int
	msgSeq           atomic.Int64 // 全局消息序号，方便排查
}

func NewService(cfg Config) *Service {
	return &Service{
		cfg:              cfg,
		dialer:           &websocket.Dialer{Proxy: http.ProxyFromEnvironment, HandshakeTimeout: 15 * time.Second},
		dedup:            newIDSet(cfg.DedupCapacity),
		latencyWarnMS:    cfg.LatencyWarnMS,
		latencyWarnCount: cfg.LatencyWarnCount,
		rttWarnMS:        cfg.RTTWarnMS,
		rttWarnCount:     cfg.RTTWarnCount,
	}
}

// Start 启动所有后台协程（worker + heartbeat），重复调用会直接返回。
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

	s.outQueue = make(chan queuedMessage, s.cfg.QueueCapacity)
	s.wg.Add(1)
	safex.SafeGo("treenews_merger", func() {
		defer s.wg.Done()
		s.mergerLoop(ctxRun)
	})

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

// Stop 通知所有后台协程退出，并等待清理结束。
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

// worker 管理单条 WebSocket 链路：负责连接、登录、启动读/心跳协程，并处理断线重连。
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
		state := newWorkerState(workerID, errCh)
		logger.GetLog().Infof("tree news worker=%d connected to %s", workerID, s.cfg.URL)
		go s.readLoop(ctx, conn, errCh, state)
		go s.heartbeatLoop(ctx, conn, errCh, state)
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
	if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
		return err
	}
	// 仅在登录成功后打印一次日志，便于确认认证流程无误
	logger.GetLog().Info("tree news login success")
	return nil
}

// readLoop 专职从 WebSocket 读取原始消息，一旦遇到异常就通过 errCh 通知上层 worker。
// readLoop 从 WebSocket 逐条读取消息，推入 outQueue，由 mergerLoop 顺序处理。
func (s *Service) readLoop(ctx context.Context, conn *websocket.Conn, errCh chan<- error, state *workerState) {
	conn.SetReadLimit(2 << 20)
	if s.cfg.PingTimeout > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(s.cfg.PingInterval + s.cfg.PingTimeout))
		conn.SetPongHandler(func(string) error {
			if s.cfg.PingTimeout > 0 {
				_ = conn.SetReadDeadline(time.Now().Add(s.cfg.PingInterval + s.cfg.PingTimeout))
			}
			if start := state.lastPing.Load(); start > 0 {
				rtt := time.Since(time.Unix(0, start)).Milliseconds()
				state.lastRTT.Store(rtt)
				if s.rttWarnMS > 0 && rtt > int64(s.rttWarnMS) {
					count := state.highRTT.Add(1)
					loggerErr.GetLog().Warnf("tree news worker=%d high rtt=%dms count=%d", state.workerID, rtt, count)
					if s.rttWarnCount > 0 && int(count) >= s.rttWarnCount {
						state.triggerReconnect(fmt.Sprintf("high rtt %dms", rtt))
					}
				} else {
					state.highRTT.Store(0)
				}
			}
			return nil
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
		msg := queuedMessage{
			data:  append([]byte(nil), data...),
			recv:  time.Now().UTC(),
			state: state,
			seq:   s.msgSeq.Add(1),
		}
		select {
		case s.outQueue <- msg:
		case <-ctx.Done():
			return
		}
	}
}

// heartbeatLoop 定期发送 ping，以检测链路健康；失败后交给 worker 重连。
func (s *Service) heartbeatLoop(ctx context.Context, conn *websocket.Conn, errCh chan<- error, state *workerState) {
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
			state.lastPing.Store(time.Now().UnixNano())
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				errCh <- err
				return
			}
		}
	}
}

// rollingLoop 控制“滚动重连”，让长时间在线的连接定期刷新。
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

// handleMessage 负责：JSON 解析 -> 去重 -> Upbit KRW 过滤 -> 回调业务 handler。
// mergerLoop 顺序消费 outQueue 中的消息，保证业务处理串行执行。
func (s *Service) mergerLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-s.outQueue:
			s.processMessage(ctx, msg)
		}
	}
}

// processMessage 执行解码、去重、延迟计算，并触发业务回调。
func (s *Service) processMessage(ctx context.Context, msg queuedMessage) {
	var payload map[string]any
	if err := json.Unmarshal(msg.data, &payload); err != nil {
		loggerErr.GetLog().Errorf("tree news json decode failed: %v", err)
		return
	}

	idRaw := toString(payload["_id"])
	trimID := strings.TrimSpace(idRaw)
	logID := trimID
	if logID == "" {
		logID = fmt.Sprintf("raw-%d", msg.seq)
	}

	event := Event{
		ID:              logID,
		Payload:         payload,
		ReceivedAt:      msg.recv,
		LatencyRawMS:    -1,
		LatencyAdjustMS: -1,
		LatencyMS:       -1,
	}
	if msg.state != nil {
		event.RTTMS = msg.state.lastRTT.Load()
	}

	if ts := toInt64(payload["time"]); ts > 0 {
		event.ServerMilli = ts
		raw := msg.recv.UnixMilli() - ts
		if raw < 0 {
			raw = 0
		}
		event.LatencyRawMS = int(raw)
		event.LatencyAdjustMS = int(raw)
		event.LatencyMS = int(raw)
	}

	logger.GetLog().Infof("tree news raw msg seq=%d id=%s latency_raw=%d latency=%d rtt=%d", msg.seq, logID, event.LatencyRawMS, event.LatencyMS, event.RTTMS)

	if msg.state != nil {
		if event.LatencyMS >= 0 && s.latencyWarnMS > 0 && event.LatencyMS > s.latencyWarnMS {
			count := msg.state.highLatency.Add(1)
			loggerErr.GetLog().Warnf("tree news worker=%d high latency=%dms count=%d id=%s", msg.state.workerID, event.LatencyMS, count, logID)
			if s.latencyWarnCount > 0 && int(count) >= s.latencyWarnCount {
				msg.state.triggerReconnect(fmt.Sprintf("high latency %dms", event.LatencyMS))
			}
		} else {
			msg.state.highLatency.Store(0)
		}
	}

	if trimID != "" && !s.dedup.Add(trimID) {
		return
	}
	if trimID != "" {
		event.ID = trimID
	}

	symbols := upbitKRWSymbols(payload)
	if len(symbols) == 0 {
		return
	}

	event.Symbols = symbols

	logger.GetLog().Infof("tree news event id=%s symbols=%v latency=%d rtt=%d", event.ID, symbols, event.LatencyMS, event.RTTMS)

	handlerMu.RLock()
	for _, h := range handlers {
		h(ctx, event)
	}
	handlerMu.RUnlock()
}

// idSet 用于维护最近 seen 消息，实现快速去重。
type idSet struct {
	mu    sync.Mutex
	order []string
	set   map[string]struct{}
	max   int
}

// newIDSet 返回一个固定容量的 idSet（超过容量时按 FIFO 淘汰旧值）。
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

type queuedMessage struct {
	data  []byte
	recv  time.Time
	state *workerState
	seq   int64
}

type workerState struct {
	workerID    int
	errCh       chan<- error
	lastPing    atomic.Int64
	lastRTT     atomic.Int64
	highRTT     atomic.Int32
	highLatency atomic.Int32
	triggered   atomic.Bool
}

func newWorkerState(workerID int, errCh chan<- error) *workerState {
	return &workerState{workerID: workerID, errCh: errCh}
}

func (ws *workerState) triggerReconnect(reason string) {
	if ws.triggered.CompareAndSwap(false, true) {
		select {
		case ws.errCh <- fmt.Errorf("worker %d: %s", ws.workerID, reason):
		default:
		}
	}
}

// toString 将任意类型转换为字符串，以兼容树新闻返回的多种 JSON 字段类型。
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

// toInt64 尝试把 JSON 字段转成 int64，用于时间戳等场景。
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

// minDuration 返回两个 duration 中的较小值。
func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
