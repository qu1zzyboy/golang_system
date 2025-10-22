package params

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/hhh500/quantGoInfra/infra/observe/log/dynamicLog"
	"github.com/hhh500/quantGoInfra/pkg/singleton"
)

const (
	// ModuleId is used in bootx registration.
	ModuleId = "params_service"

	defaultFGI = 50
)

var (
	serviceSingleton = singleton.NewSingleton(func() *Service {
		return &Service{
			cfg: defaultConfig(),
		}
	})

	log      = dynamicLog.Log
	logError = dynamicLog.Error
)

// Service is the central coordinator that aggregates market data sources
// (BTC returns, Fear&Greed index, Binance OI) and exposes a synchronous
// Compute method for strategy usage.
type Service struct {
	cfg           Config
	startStopOnce sync.Once
	stopOnce      sync.Once

	btc *BTCMetrics
	fgi *FGIPoller
	oi  *OIStore
}

func GetService() *Service {
	return serviceSingleton.Get()
}

// Config contains tunables for the background data providers.
type Config struct {
	BTC BTCConfig
	FGI FGIConfig
	OI  OIConfig
}

func defaultConfig() Config {
	return Config{
		BTC: BTCConfig{
			Symbol:            "BTCUSDT",
			H1WindowSize:      240,
			H1RegularPullSec:  600,
			H1EdgePreSec:      20,
			H1EdgePostSec:     20,
			M1PullSec:         45,
			RequestTimeout:    5 * time.Second,
			StartReadyTimeout: 20 * time.Second,
		},
		FGI: FGIConfig{
			Interval:         30 * time.Second,
			RequestTimeout:   5 * time.Second,
			StartReadyWait:   10 * time.Second,
			DefaultFGIValue:  defaultFGI,
			FallbackFGIValue: defaultFGI,
		},
		OI: OIConfig{
			RedisKey:     "BN_OPEN_INTEREST",
			RefreshEvery: 60 * time.Second,
			Timeout:      1*time.Second + 500*time.Millisecond,
		},
	}
}

// Start boots all background providers. It is idempotent.
func (s *Service) Start(ctx context.Context) error {
	var startErr error
	s.startStopOnce.Do(func() {
		s.btc = NewBTCMetrics(s.cfg.BTC)
		s.fgi = NewFGIPoller(s.cfg.FGI)
		s.oi = NewOIStore(s.cfg.OI)

		if err := s.btc.Start(ctx); err != nil {
			startErr = err
			return
		}
		if err := s.fgi.Start(ctx); err != nil {
			startErr = err
			return
		}
		if err := s.oi.Start(ctx); err != nil {
			startErr = err
			return
		}
	})
	return startErr
}

// Stop gracefully stops background goroutines. It is safe to call multiple times.
func (s *Service) Stop(ctx context.Context) error {
	var stopErr error
	s.stopOnce.Do(func() {
		if s.oi != nil {
			if err := s.oi.Stop(ctx); err != nil && stopErr == nil {
				stopErr = err
			}
		}
		if s.fgi != nil {
			if err := s.fgi.Stop(ctx); err != nil && stopErr == nil {
				stopErr = err
			}
		}
		if s.btc != nil {
			if err := s.btc.Stop(ctx); err != nil && stopErr == nil {
				stopErr = err
			}
		}
	})
	return stopErr
}

// ComputeRequest describes the inputs required for gain/twap evaluation.
type ComputeRequest struct {
	MarketCapM float64
	IsMeme     bool
	SymbolName string
}

// Diagnostics captures metadata used for observability. Fields largely mirror
// the Python implementation to keep parity during the migration.
type Diagnostics struct {
	OI               *float64
	OITimestamp      *int64
	S                *float64
	SNorm            *float64
	GainBase         float64
	GainOIAdd        float64
	GainFinal        float64
	TwapBase         float64
	TwapOIAdd        float64
	TwapFinal        float64
	FGI              int
	BTC1D            float64
	BTC7D            float64
	AsOf             string
	StalenessSeconds int
}

// ComputeResponse is the outcome returned to strategy callers.
type ComputeResponse struct {
	GainPct float64
	TwapSec float64
	Diag    Diagnostics
}

// Compute replicates the Python strategy logic:
//  1. fetch BTC metrics + FGI (with staleness guarding)
//  2. evaluate baseline gain/twap from market-cap buckets
//  3. add OI-based adjustments
//  4. clip by bucket bounds
func (s *Service) Compute(ctx context.Context, req ComputeRequest) (ComputeResponse, error) {
	if ctx != nil {
		select {
		case <-ctx.Done():
			return ComputeResponse{}, ctx.Err()
		default:
		}
	}
	if s.btc == nil || s.fgi == nil || s.oi == nil {
		return ComputeResponse{}, errors.New("params service not started")
	}

	snapshot := s.btc.Snapshot()
	fgiValue := s.fgi.GetValue()
	if fgiValue <= 0 {
		fgiValue = defaultFGI
	}

	btc1d := snapshot.BTC1D
	if math.IsNaN(btc1d) {
		btc1d = 0
	}
	btc7d := snapshot.BTC7D
	if math.IsNaN(btc7d) {
		btc7d = 0
	}
	gainBase, twapBase := expectedSplitGainAndTwapDuration(req.MarketCapM, float64(fgiValue), btc1d, btc7d, req.IsMeme)

	var (
		oiValue *float64
		oiTime  *int64
	)
	if req.SymbolName != "" {
		if record, ok := s.oi.Get(req.SymbolName); ok {
			if record.OpenInterest != nil {
				oiValue = record.OpenInterest
			}
			if record.Timestamp > 0 {
				ts := record.Timestamp
				oiTime = &ts
			}
		}
	}

	gainAdd, twapAdd, strength, norm := computeOIContribs(oiValue, req.MarketCapM)
	gainFinal, gainMin, gainMax := clipGain(req.MarketCapM, gainBase+gainAdd)
	twapFinal, twapMin, twapMax := clipTwap(req.MarketCapM, twapBase+twapAdd)

	diag := Diagnostics{
		GainBase:         gainBase,
		GainOIAdd:        gainAdd,
		GainFinal:        gainFinal,
		TwapBase:         twapBase,
		TwapOIAdd:        twapAdd,
		TwapFinal:        twapFinal,
		FGI:              fgiValue,
		BTC1D:            btc1d,
		BTC7D:            btc7d,
		AsOf:             snapshot.AsOf,
		StalenessSeconds: snapshot.StalenessSeconds,
	}
	if oiValue != nil {
		diag.OI = oiValue
	}
	if oiTime != nil {
		diag.OITimestamp = oiTime
	}
	if strength != nil {
		diag.S = strength
	}
	if norm != nil {
		diag.SNorm = norm
	}
	// record the clip bands for debugging
	diag.GainFinal = clampFloat(gainFinal, gainMin, gainMax)
	diag.TwapFinal = clampFloat(twapFinal, twapMin, twapMax)

	resp := ComputeResponse{
		GainPct: diag.GainFinal,
		TwapSec: diag.TwapFinal,
		Diag:    diag,
	}
	return resp, nil
}

func clampFloat(value, lower, upper float64) float64 {
	if value < lower {
		return lower
	}
	if value > upper {
		return upper
	}
	return value
}
