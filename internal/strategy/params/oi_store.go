package params

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hhh500/quantGoInfra/infra/redisx"
	"github.com/hhh500/quantGoInfra/infra/redisx/redisConfig"
	"github.com/hhh500/quantGoInfra/infra/safex"
)

// OIConfig describes how open-interest snapshots are refreshed.
type OIConfig struct {
	RedisKey     string
	RefreshEvery time.Duration
	Timeout      time.Duration
}

// OIRecord is the normalized view per symbol.
type OIRecord struct {
	Symbol       string
	OpenInterest *float64
	OpenQty      *float64
	Timestamp    int64
	TimeStr      string
}

type OIStore struct {
	cfg       OIConfig
	startOnce sync.Once
	stopOnce  sync.Once
	cancel    context.CancelFunc

	mu     sync.RWMutex
	data   map[string]OIRecord
	client *redis.Client
}

func NewOIStore(cfg OIConfig) *OIStore {
	if cfg.RedisKey == "" {
		cfg.RedisKey = "BN_OPEN_INTEREST"
	}
	if cfg.RefreshEvery <= 0 {
		cfg.RefreshEvery = 60 * time.Second
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 1500 * time.Millisecond
	}
	return &OIStore{
		cfg:  cfg,
		data: make(map[string]OIRecord),
	}
}

func (s *OIStore) Start(ctx context.Context) error {
	var startErr error
	s.startOnce.Do(func() {
		client, err := redisx.LoadClient(redisConfig.CONFIG_ALL_KEY)
		if err != nil {
			startErr = fmt.Errorf("load redis client failed: %w", err)
			return
		}
		s.client = client

		ctxRun, cancel := context.WithCancel(context.Background())
		s.cancel = cancel
		if ctx != nil {
			safex.SafeGo("oi_store_cancel", func() {
				select {
				case <-ctx.Done():
					cancel()
				case <-ctxRun.Done():
				}
			})
		}

		if err := s.refresh(ctxRun); err != nil {
			startErr = err
			return
		}

		safex.SafeGo("oi_store_loop", func() {
			s.loop(ctxRun)
		})
	})
	return startErr
}

func (s *OIStore) Stop(ctx context.Context) error {
	s.stopOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
	})
	return nil
}

func (s *OIStore) loop(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.RefreshEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.refresh(ctx); err != nil {
				// keep previous data; log once
				logError.GetLog().Errorf("oi refresh failed: %v", err)
			}
		}
	}
}

// Get returns the latest OI record for a symbol.
func (s *OIStore) Get(symbol string) (OIRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rec, ok := s.data[symbol]
	return rec, ok
}

func (s *OIStore) refresh(ctx context.Context) error {
	if s.client == nil {
		return errors.New("redis client not initialized")
	}
	ctx, cancel := context.WithTimeout(ctx, s.cfg.Timeout)
	defer cancel()

	res, err := s.client.HGetAll(ctx, s.cfg.RedisKey).Result()
	if err != nil {
		return fmt.Errorf("redis HGetAll key=%s: %w", s.cfg.RedisKey, err)
	}
	if len(res) == 0 {
		// fallback to plain GET big JSON
		str, err := s.client.Get(ctx, s.cfg.RedisKey).Result()
		if err != nil {
			return fmt.Errorf("redis GET key=%s: %w", s.cfg.RedisKey, err)
		}
		return s.refreshFromJSON(str)
	}
	return s.refreshFromHash(res)
}

func (s *OIStore) refreshFromHash(raw map[string]string) error {
	updated := make(map[string]OIRecord, len(raw))
	for _, v := range raw {
		rec, err := parseOIRecord(v)
		if err != nil {
			continue
		}
		if rec.Symbol == "" {
			continue
		}
		if existing, ok := updated[rec.Symbol]; ok {
			if rec.Timestamp > existing.Timestamp {
				updated[rec.Symbol] = rec
			}
		} else {
			updated[rec.Symbol] = rec
		}
	}
	s.mu.Lock()
	s.data = updated
	s.mu.Unlock()
	return nil
}

func (s *OIStore) refreshFromJSON(raw string) error {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		return err
	}
	updated := make(map[string]OIRecord, len(obj))
	for _, payload := range obj {
		rec, err := parseOIRecordBytes(payload)
		if err != nil {
			continue
		}
		if rec.Symbol == "" {
			continue
		}
		if existing, ok := updated[rec.Symbol]; ok {
			if rec.Timestamp > existing.Timestamp {
				updated[rec.Symbol] = rec
			}
		} else {
			updated[rec.Symbol] = rec
		}
	}
	s.mu.Lock()
	s.data = updated
	s.mu.Unlock()
	return nil
}

func parseOIRecord(raw string) (OIRecord, error) {
	return parseOIRecordBytes(json.RawMessage(raw))
}

func parseOIRecordBytes(raw json.RawMessage) (OIRecord, error) {
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return OIRecord{}, err
	}
	rec := OIRecord{}
	if v, ok := obj["symbol"].(string); ok {
		rec.Symbol = v
	}
	if oi, ok := numericPointer(obj["open_interest"]); ok {
		rec.OpenInterest = oi
	}
	if oq, ok := numericPointer(obj["open_qty"]); ok {
		rec.OpenQty = oq
	}
	if t, ok := toInt64(obj["time"]); ok {
		rec.Timestamp = t
	}
	if s, ok := obj["time_str"].(string); ok {
		rec.TimeStr = s
	}
	return rec, nil
}

func numericPointer(v any) (*float64, bool) {
	value, ok := toFloat64(v)
	if !ok {
		return nil, false
	}
	return &value, true
}
