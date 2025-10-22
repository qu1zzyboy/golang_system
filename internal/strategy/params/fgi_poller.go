package params

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/hhh500/quantGoInfra/infra/safex"
)

// FGIConfig 描述恐惧贪婪指数轮询的行为参数。
type FGIConfig struct {
	Interval         time.Duration
	RequestTimeout   time.Duration
	StartReadyWait   time.Duration
	DefaultFGIValue  int
	FallbackFGIValue int
}

type fgiState struct {
	Value           int
	Classification  string
	Timestamp       int64
	TimeUntilUpdate int64
}

// FGIPoller 周期性拉取 alternative.me 的指数，并缓存最新结果。
type FGIPoller struct {
	cfg       FGIConfig
	client    *http.Client
	startOnce sync.Once
	stopOnce  sync.Once
	cancel    context.CancelFunc

	mu    sync.RWMutex
	state fgiState
	ready chan struct{}
}

func NewFGIPoller(cfg FGIConfig) *FGIPoller {
	if cfg.Interval <= 0 {
		cfg.Interval = 30 * time.Second
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 5 * time.Second
	}
	if cfg.StartReadyWait <= 0 {
		cfg.StartReadyWait = 10 * time.Second
	}
	if cfg.DefaultFGIValue == 0 {
		cfg.DefaultFGIValue = 50
	}
	if cfg.FallbackFGIValue == 0 {
		cfg.FallbackFGIValue = cfg.DefaultFGIValue
	}
	return &FGIPoller{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.RequestTimeout,
		},
		ready: make(chan struct{}),
	}
}

func (p *FGIPoller) Start(ctx context.Context) error {
	var startErr error
	p.startOnce.Do(func() {
		ctxRun, cancel := context.WithCancel(context.Background())
		p.cancel = cancel
		if ctx != nil {
			safex.SafeGo("fgi_cancel", func() {
				select {
				case <-ctx.Done():
					cancel()
				case <-ctxRun.Done():
				}
			})
		}

		if err := p.fetchOnce(ctxRun); err != nil {
			startErr = fmt.Errorf("initial fgi fetch failed: %w", err)
		}
		close(p.ready)

		if startErr == nil {
			safex.SafeGo("fgi_loop", func() {
				p.loop(ctxRun)
			})
		}
	})
	return startErr
}

func (p *FGIPoller) Stop(ctx context.Context) error {
	p.stopOnce.Do(func() {
		if p.cancel != nil {
			p.cancel()
		}
	})
	return nil
}

func (p *FGIPoller) loop(ctx context.Context) {
	interval := p.cfg.Interval
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.fetchOnce(ctx); err != nil {
				// 退避策略：每次翻倍直至封顶 30 秒。
				interval = time.Duration(minInt(int(interval.Seconds()*2), 30)) * time.Second
				ticker.Reset(interval)
			} else {
				if interval != p.cfg.Interval {
					interval = p.cfg.Interval
					ticker.Reset(interval)
				}
			}
		}
	}
}

func (p *FGIPoller) fetchOnce(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.alternative.me/fng/?limit=2", nil)
	if err != nil {
		return err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("fgi status %d", resp.StatusCode)
	}
	var payload struct {
		Data []struct {
			Value               string `json:"value"`
			ValueClassification string `json:"value_classification"`
			Timestamp           string `json:"timestamp"`
			TimeUntilUpdate     string `json:"time_until_update"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return err
	}
	if len(payload.Data) == 0 {
		return errors.New("fgi payload empty")
	}
	item := payload.Data[0]

	value, err := strconv.Atoi(item.Value)
	if err != nil {
		return fmt.Errorf("invalid fgi value: %w", err)
	}
	timestamp, err := strconv.ParseInt(item.Timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid fgi timestamp: %w", err)
	}
	var ttu int64
	if item.TimeUntilUpdate != "" {
		if v, err := strconv.ParseInt(item.TimeUntilUpdate, 10, 64); err == nil {
			ttu = v
		}
	}

	newState := fgiState{
		Value:           value,
		Classification:  item.ValueClassification,
		Timestamp:       timestamp,
		TimeUntilUpdate: ttu,
	}
	p.mu.Lock()
	p.state = newState
	p.mu.Unlock()
	return nil
}

// GetValue 返回最新指数，若不可用则回退至默认值。
func (p *FGIPoller) GetValue() int {
	select {
	case <-p.ready:
	case <-time.After(p.cfg.StartReadyWait):
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.state.Value == 0 {
		return p.cfg.FallbackFGIValue
	}
	return p.state.Value
}
