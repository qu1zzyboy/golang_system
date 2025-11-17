package treenews

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"upbitBnServer/internal/infra/observe/log/logCfg"

	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/observe/log/staticLog"
	"upbitBnServer/pkg/singleton"
	"upbitBnServer/pkg/utils/timeUtils"

	"github.com/gorilla/websocket"
)

var (
	treeLogConfig = staticLog.Config{
		NeedErrorHook: true,
		FileDir:       "tree_news_log",
		DateStr:       timeUtils.GetNowDateStr(),
		FileName:      "tree_news",
		Level:         logCfg.INFO_LEVEL,
	}
	treeErrConfig = staticLog.Config{
		NeedErrorHook: false,
		FileDir:       "tree_news_log",
		DateStr:       timeUtils.GetNowDateStr(),
		FileName:      "tree_news_error",
		Level:         logCfg.ERROR_LEVEL,
	}
	logger    = dynamicLog.NewDynamicLogger(treeLogConfig)
	loggerErr = dynamicLog.NewDynamicLogger(treeErrConfig)
)

var (
	bnSingleton = singleton.NewSingleton(func() *Service {
		return NewService(defaultConfig())
	})
)

func GetService() *Service {
	return bnSingleton.Get()
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

// Service 管理 WebSocket 工作者并将事件分发给回调。
// Service 维护 WebSocket 连接、读写协程和消息去重等核心状态。
type Service struct {
	cfg              Config
	dedup            *idSet // dedup 保存最近的消息 id，防止重复处理
	latencyWarnMS    int
	latencyWarnCount int
	rttWarnMS        int
	rttWarnCount     int
	msgSeq           atomic.Int64 // 全局消息序号，方便排查
}

func NewService(cfg Config) *Service {
	return &Service{
		cfg:              cfg,
		dedup:            newIDSet(cfg.DedupCapacity),
		latencyWarnMS:    cfg.LatencyWarnMS,
		latencyWarnCount: cfg.LatencyWarnCount,
		rttWarnMS:        cfg.RTTWarnMS,
		rttWarnCount:     cfg.RTTWarnCount,
	}
}

// Start 启动所有后台协程（worker + heartbeat），重复调用会直接返回。
func (s *Service) Start(ctx context.Context) error {
	s.worker(ctxRun)
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	return nil
}

// worker 管理单条 WebSocket 链路：负责连接、登录、启动读/心跳协程，并处理断线重连。
func (s *Service) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			loggerErr.GetLog().Errorf("tree news dial failed: %v", err)
			continue
		}

		if err := s.login(conn); err != nil {
			loggerErr.GetLog().Errorf("tree news login failed: %v", err)
			conn.Close()
		}
		continue
	}
	go s.readLoop(ctx, conn, errCh, state)
}

func (s *Service) login(conn *websocket.Conn) error {
	message := fmt.Sprintf("login %s", s.cfg.APIKey)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
		return err
	}
	logger.GetLog().Info("tree news login success")
	return nil
}

// readLoop 专职从 WebSocket 读取原始消息，一旦遇到异常就通过 errCh 通知上层 worker。
// readLoop 从 WebSocket 逐条读取消息，推入 outQueue，由 mergerLoop 顺序处理。
func (s *Service) readLoop(ctx context.Context, conn *websocket.Conn, state *workerState) {
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		msg := queuedMessage{
			data:  data,
			recv:  time.Now().UTC(),
			state: state,
			seq:   s.msgSeq.Add(1),
		}
		s.processMessage(msg)
	}
}

// handleMessage 负责：JSON 解析 -> 去重 -> Upbit KRW 过滤 -> 回调业务 handler。
// processMessage 执行解码、去重、延迟计算，并触发业务回调。
func (s *Service) processMessage(msg queuedMessage) {
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

	source := strings.ToLower(toString(payload["source"]))
	if source == "" {
		source = strings.ToLower(toString(payload["type"]))
	}

	if ts := toInt64(payload["time"]); ts > 0 {
		event.ServerMilli = ts
		raw := msg.recv.UnixMilli() - ts
		if raw < 0 {
			raw = 0
		}
		adjust := raw
		if source == "blogs" {
			adjust -= 5000
			if adjust < 0 {
				adjust = 0
			}
		}
		event.LatencyRawMS = int(raw)
		event.LatencyAdjustMS = int(adjust)
		event.LatencyMS = int(adjust)
	}

	logger.GetLog().Infof("tree news raw msg seq=%d id=%s latency_raw=%d latency=%d rtt=%d", msg.seq, logID, event.LatencyRawMS, event.LatencyMS, event.RTTMS)

	if msg.state != nil {
		if event.LatencyMS >= 0 && s.latencyWarnMS > 0 && event.LatencyMS > s.latencyWarnMS {
			count := msg.state.highLatency.Add(1)
			loggerErr.GetLog().Warnf("tree news worker=%d high latency=%dms count=%d id=%s", msg.state.workerID, event.LatencyMS, count, logID)
			if s.latencyWarnCount > 0 && int(count) >= s.latencyWarnCount {
				// msg.state.triggerReconnect(fmt.Sprintf("high latency %dms", event.LatencyMS))
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
	lastPing    atomic.Int64
	lastRTT     atomic.Int64
	highRTT     atomic.Int32
	highLatency atomic.Int32
	triggered   atomic.Bool
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
