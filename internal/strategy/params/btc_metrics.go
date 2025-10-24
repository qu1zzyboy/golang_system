package params

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"upbitBnServer/internal/infra/safex"
)

// BTCConfig 对应 Python 版本的可调参数。
type BTCConfig struct {
	Symbol            string
	H1WindowSize      int
	H1RegularPullSec  int
	H1EdgePreSec      int
	H1EdgePostSec     int
	M1PullSec         int
	RequestTimeout    time.Duration
	StartReadyTimeout time.Duration
}

type h1Bar struct {
	CloseTime time.Time
	Close     float64
}

// BTCMetrics 周期性获取币安合约 K 线，维护 24h/7d 收益计算所需的历史数据，并保证并发安全。
type BTCMetrics struct {
	cfg       BTCConfig
	client    *http.Client
	startOnce sync.Once
	stopOnce  sync.Once
	cancel    context.CancelFunc
	mu        sync.RWMutex
	h1Bars    []h1Bar
	nowPrice  float64
	nowAsOf   time.Time
	lastH1Err error
	lastM1Err error
}

// BTCSnapshot 供调用方读取最新指标。
type BTCSnapshot struct {
	BTC1D            float64
	BTC7D            float64
	AsOf             string
	StalenessSeconds int
}

func NewBTCMetrics(cfg BTCConfig) *BTCMetrics {
	if cfg.Symbol == "" {
		cfg.Symbol = "BTCUSDT"
	}
	if cfg.H1WindowSize < 180 {
		cfg.H1WindowSize = 180
	}
	if cfg.H1RegularPullSec < 60 {
		cfg.H1RegularPullSec = 60
	}
	if cfg.M1PullSec < 10 {
		cfg.M1PullSec = 10
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 5 * time.Second
	}
	if cfg.StartReadyTimeout <= 0 {
		cfg.StartReadyTimeout = 20 * time.Second
	}

	return &BTCMetrics{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.RequestTimeout,
		},
	}
}

func (m *BTCMetrics) Start(ctx context.Context) error {
	var startErr error
	m.startOnce.Do(func() {
		ctxRun, cancel := context.WithCancel(context.Background())
		m.cancel = cancel
		if ctx != nil {
			safex.SafeGo("btc_metrics_cancel", func() {
				select {
				case <-ctx.Done():
					cancel()
				case <-ctxRun.Done():
				}
			})
		}

		if err := m.bootstrapH1(ctxRun); err != nil {
			startErr = fmt.Errorf("btc bootstrap failed: %w", err)
			return
		}
		if err := m.pullM1(ctxRun); err != nil {
			startErr = fmt.Errorf("btc initial m1 failed: %w", err)
			return
		}

		safex.SafeGo("btc_metrics_loop", func() {
			m.loop(ctxRun)
		})

		if ready := m.waitUntilReady(); !ready {
			startErr = errors.New("btc metrics not ready within timeout")
		}
	})
	return startErr
}

func (m *BTCMetrics) Stop(ctx context.Context) error {
	var stopErr error
	m.stopOnce.Do(func() {
		if m.cancel != nil {
			m.cancel()
		}
	})
	return stopErr
}

func (m *BTCMetrics) Snapshot() BTCSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	asof := ""
	staleness := -1
	if !m.nowAsOf.IsZero() {
		asof = m.nowAsOf.UTC().Format(time.RFC3339)
		staleness = int(time.Since(m.nowAsOf).Seconds())
	}

	if math.IsNaN(m.nowPrice) || len(m.h1Bars) < 2 {
		return BTCSnapshot{
			BTC1D:            math.NaN(),
			BTC7D:            math.NaN(),
			AsOf:             asof,
			StalenessSeconds: staleness,
		}
	}

	nowMs := m.nowAsOf.UTC().UnixMilli()
	bars := append([]h1Bar(nil), m.h1Bars...)
	return BTCSnapshot{
		BTC1D:            m.rollingReturn(bars, m.nowPrice, nowMs, 24),
		BTC7D:            m.rollingReturn(bars, m.nowPrice, nowMs, 24*7),
		AsOf:             asof,
		StalenessSeconds: staleness,
	}
}

func (m *BTCMetrics) waitUntilReady() bool {
	deadline := time.After(m.cfg.StartReadyTimeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			return false
		case <-ticker.C:
			m.mu.RLock()
			ok := len(m.h1Bars) >= 170 && !m.nowAsOf.IsZero()
			m.mu.RUnlock()
			if ok {
				return true
			}
		}
	}
}

func (m *BTCMetrics) loop(ctx context.Context) {
	nextH1Regular := time.Now().Add(time.Duration(m.cfg.H1RegularPullSec) * time.Second)
	nextM1 := time.Now().Add(time.Duration(m.cfg.M1PullSec) * time.Second)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			m.tryEdgePull(ctx, now)

			if now.After(nextH1Regular) {
				if err := m.pullH1Incremental(ctx, 5); err != nil {
					m.setLastH1Err(err)
				}
				nextH1Regular = now.Add(time.Duration(m.cfg.H1RegularPullSec) * time.Second)
			}

			if now.After(nextM1) {
				if err := m.pullM1(ctx); err != nil {
					m.setLastM1Err(err)
				}
				nextM1 = now.Add(time.Duration(m.cfg.M1PullSec) * time.Second)
			}
		}
	}
}

func (m *BTCMetrics) tryEdgePull(ctx context.Context, now time.Time) {
	if m.cfg.H1EdgePreSec <= 0 && m.cfg.H1EdgePostSec <= 0 {
		return
	}
	secIntoHour := now.UTC().Minute()*60 + now.UTC().Second()
	if 3600-secIntoHour <= m.cfg.H1EdgePreSec || secIntoHour <= m.cfg.H1EdgePostSec {
		if err := m.pullH1Incremental(ctx, 5); err != nil {
			m.setLastH1Err(err)
		}
	}
}

func (m *BTCMetrics) setLastH1Err(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastH1Err = err
}

func (m *BTCMetrics) setLastM1Err(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastM1Err = err
}

func (m *BTCMetrics) bootstrapH1(ctx context.Context) error {
	kl, err := m.fetchKlines(ctx, "1h", minInt(m.cfg.H1WindowSize, 1000))
	if err != nil {
		return err
	}
	bars := klinesToH1(kl)
	m.mu.Lock()
	m.h1Bars = trimH1Bars(bars, m.cfg.H1WindowSize)
	m.mu.Unlock()
	return nil
}

func (m *BTCMetrics) pullH1Incremental(ctx context.Context, limit int) error {
	kl, err := m.fetchKlines(ctx, "1h", limit)
	if err != nil {
		return err
	}
	bars := klinesToH1(kl)
	if len(bars) == 0 {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.h1Bars) == 0 {
		m.h1Bars = trimH1Bars(bars, m.cfg.H1WindowSize)
		return nil
	}

	last := m.h1Bars[len(m.h1Bars)-1].CloseTime
	for _, b := range bars {
		if b.CloseTime.After(last) {
			m.h1Bars = append(m.h1Bars, b)
		}
	}
	m.h1Bars = trimH1Bars(m.h1Bars, m.cfg.H1WindowSize)
	return nil
}

func (m *BTCMetrics) pullM1(ctx context.Context) error {
	kl, err := m.fetchKlines(ctx, "1m", 1)
	if err != nil {
		return err
	}
	if len(kl) == 0 {
		return nil
	}
	record := kl[len(kl)-1]
	closePrice, closeTime, err := parseClose(record)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.nowPrice = closePrice
	m.nowAsOf = closeTime
	m.mu.Unlock()
	return nil
}

func (m *BTCMetrics) fetchKlines(ctx context.Context, interval string, limit int) ([][]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://fapi.binance.com/fapi/v1/klines", nil)
	if err != nil {
		return nil, err
	}
	query := req.URL.Query()
	query.Set("symbol", m.cfg.Symbol)
	query.Set("interval", interval)
	query.Set("limit", fmt.Sprintf("%d", limit))
	req.URL.RawQuery = query.Encode()

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch klines status %d", resp.StatusCode)
	}

	var data [][]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func klinesToH1(klines [][]any) []h1Bar {
	res := make([]h1Bar, 0, len(klines))
	for _, k := range klines {
		if len(k) < 7 {
			continue
		}
		closeValue, ok := toFloat64(k[4])
		if !ok {
			continue
		}
		closeTimeMs, ok := toInt64(k[6])
		if !ok {
			continue
		}
		res = append(res, h1Bar{
			CloseTime: time.UnixMilli(closeTimeMs).UTC(),
			Close:     closeValue,
		})
	}
	return res
}

func parseClose(kline []any) (float64, time.Time, error) {
	if len(kline) < 7 {
		return 0, time.Time{}, errors.New("kline too short")
	}
	price, ok := toFloat64(kline[4])
	if !ok {
		return 0, time.Time{}, errors.New("invalid price")
	}
	tsMs, ok := toInt64(kline[6])
	if !ok {
		return 0, time.Time{}, errors.New("invalid timestamp")
	}
	return price, time.UnixMilli(tsMs).UTC(), nil
}

func trimH1Bars(bars []h1Bar, window int) []h1Bar {
	if len(bars) <= window {
		return append([]h1Bar(nil), bars...)
	}
	return append([]h1Bar(nil), bars[len(bars)-window:]...)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m *BTCMetrics) rollingReturn(bars []h1Bar, nowPrice float64, nowMs int64, hours int) float64 {
	ref := nowMs - int64(hours)*3600*1000
	refPrice := interpRefPrice(bars, ref)
	if refPrice <= 0 || math.IsNaN(refPrice) {
		return math.NaN()
	}
	return (nowPrice - refPrice) / refPrice * 100.0
}

func interpRefPrice(bars []h1Bar, refMs int64) float64 {
	if len(bars) < 2 {
		return math.NaN()
	}
	if refMs < bars[0].CloseTime.UnixMilli() || refMs > bars[len(bars)-1].CloseTime.UnixMilli() {
		return math.NaN()
	}

	lo, hi := 0, len(bars)-1
	for lo < hi {
		mid := (lo + hi) / 2
		if bars[mid].CloseTime.UnixMilli() < refMs {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	idx := lo
	if bars[idx].CloseTime.UnixMilli() == refMs {
		return bars[idx].Close
	}
	if idx == 0 {
		return bars[0].Close
	}
	left := bars[idx-1]
	right := bars[idx]
	t0 := left.CloseTime.UnixMilli()
	t1 := right.CloseTime.UnixMilli()
	if t1 == t0 {
		return left.Close
	}
	alpha := float64(refMs-t0) / float64(t1-t0)
	return left.Close + alpha*(right.Close-left.Close)
}

func toFloat64(v any) (float64, bool) {
	switch vv := v.(type) {
	case float64:
		return vv, true
	case string:
		val, err := strconv.ParseFloat(vv, 64)
		if err != nil {
			return 0, false
		}
		return val, true
	case json.Number:
		val, err := vv.Float64()
		if err != nil {
			return 0, false
		}
		return val, true
	default:
		return 0, false
	}
}

func toInt64(v any) (int64, bool) {
	switch vv := v.(type) {
	case float64:
		return int64(vv), true
	case int64:
		return vv, true
	case int:
		return int64(vv), true
	case json.Number:
		val, err := vv.Int64()
		if err != nil {
			return 0, false
		}
		return val, true
	case string:
		val, err := strconv.ParseInt(vv, 10, 64)
		if err != nil {
			return 0, false
		}
		return val, true
	default:
		return 0, false
	}
}
