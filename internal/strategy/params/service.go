package params

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/pkg/singleton"
)

const (
	// ModuleId 用于在 bootx 中注册该模块。
	ModuleId = "params_service"

	defaultFGI = 50
)

// BTCProvider 定义 BTC 指标数据源接口，用于支持 mock。
type BTCProvider interface {
	Start(context.Context) error
	Stop(context.Context) error
	Snapshot() BTCSnapshot
}

// FGIProvider 定义恐惧贪婪指数数据源接口。
type FGIProvider interface {
	Start(context.Context) error
	Stop(context.Context) error
	GetValue() int
}

// OIProvider 定义 OI 数据源接口。
type OIProvider interface {
	Start(context.Context) error
	Stop(context.Context) error
	Get(symbol string) (OIRecord, bool)
}

type providerFactories struct {
	newBTC func(BTCConfig) BTCProvider
	newFGI func(FGIConfig) FGIProvider
	newOI  func(OIConfig) OIProvider
}

var (
	serviceSingleton = singleton.NewSingleton(func() *Service {
		return newService(defaultConfig(), defaultFactories())
	})

	logError = dynamicLog.Error
)

// Service 为参数计算核心，负责整合 BTC 收益、恐惧贪婪指数与币安 OI，并提供同步计算接口。
type Service struct {
	mu        sync.Mutex
	cfg       Config
	factories providerFactories

	btc     BTCProvider
	fgi     FGIProvider
	oi      OIProvider
	started bool
	err     error
}

func newService(cfg Config, factories providerFactories) *Service {
	return &Service{
		cfg:       cfg,
		factories: factories,
	}
}

func defaultFactories() providerFactories {
	return providerFactories{
		newBTC: func(cfg BTCConfig) BTCProvider { return NewBTCMetrics(cfg) },
		newFGI: func(cfg FGIConfig) FGIProvider { return NewFGIPoller(cfg) },
		newOI:  func(cfg OIConfig) OIProvider { return NewOIStore(cfg) },
	}
}

func GetService() *Service {
	return serviceSingleton.Get()
}

// Config 定义各后台数据源的配置项。
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

// NewWithProviders 允许在测试中注入自定义 Provider。
func NewWithProviders(cfg Config, btc BTCProvider, fgi FGIProvider, oi OIProvider) *Service {
	return &Service{
		cfg:       cfg,
		factories: defaultFactories(),
		btc:       btc,
		fgi:       fgi,
		oi:        oi,
	}
}

// SetConfig 在启动前调整配置。
func (s *Service) SetConfig(cfg Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return errors.New("params service already started")
	}
	s.cfg = cfg
	return nil
}

// SetProviderFactories 在启动前替换默认工厂，方便注入 mock。
func (s *Service) SetProviderFactories(btc func(BTCConfig) BTCProvider, fgi func(FGIConfig) FGIProvider, oi func(OIConfig) OIProvider) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return errors.New("params service already started")
	}
	if btc != nil {
		s.factories.newBTC = btc
	}
	if fgi != nil {
		s.factories.newFGI = fgi
	}
	if oi != nil {
		s.factories.newOI = oi
	}
	return nil
}

// SetProviders 在启动前直接注入实例。
func (s *Service) SetProviders(btc BTCProvider, fgi FGIProvider, oi OIProvider) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return errors.New("params service already started")
	}
	s.btc = btc
	s.fgi = fgi
	s.oi = oi
	return nil
}

// Start 启动所有后台任务。
func (s *Service) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.started {
		err := s.err
		s.mu.Unlock()
		return err
	}
	if s.btc == nil {
		s.btc = s.factories.newBTC(s.cfg.BTC)
	}
	if s.fgi == nil {
		s.fgi = s.factories.newFGI(s.cfg.FGI)
	}
	if s.oi == nil {
		s.oi = s.factories.newOI(s.cfg.OI)
	}
	s.mu.Unlock()

	if err := s.btc.Start(ctx); err != nil {
		s.setState(false, err)
		return err
	}
	if err := s.fgi.Start(ctx); err != nil {
		s.btc.Stop(ctx)
		s.setState(false, err)
		return err
	}
	if err := s.oi.Start(ctx); err != nil {
		s.fgi.Stop(ctx)
		s.btc.Stop(ctx)
		s.setState(false, err)
		return err
	}
	s.setState(true, nil)
	return nil
}

// Stop 优雅关闭后台协程。
func (s *Service) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return nil
	}

	btc := s.btc
	fgi := s.fgi
	oi := s.oi
	s.started = false
	s.err = nil
	s.mu.Unlock()

	var firstErr error
	if oi != nil {
		if err := oi.Stop(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if fgi != nil {
		if err := fgi.Stop(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if btc != nil {
		if err := btc.Stop(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (s *Service) setState(started bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.started = started
	s.err = err
}

// ComputeRequest 描述计算止盈与 TWAP 所需的入参。
type ComputeRequest struct {
	MarketCapM float64
	IsMeme     bool
	SymbolName string
}

// Diagnostics 记录观测字段，与 Python 版本保持一致便于对齐。
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

// ComputeResponse 是返回给策略层的结果。
type ComputeResponse struct {
	GainPct float64
	TwapSec float64
	Diag    Diagnostics
}

// Compute 复刻 Python 的策略流程：
//  1. 获取 BTC 指标与 FGI，并处理数据陈旧；
//  2. 根据市值分桶得到基准 gain/twap；
//  3. 叠加 OI 修正项；
//  4. 按分桶上下限裁剪结果。
func (s *Service) Compute(ctx context.Context, req ComputeRequest) (ComputeResponse, error) {
	if ctx != nil {
		select {
		case <-ctx.Done():
			return ComputeResponse{}, ctx.Err()
		default:
		}
	}

	s.mu.Lock()
	btc := s.btc
	fgi := s.fgi
	oi := s.oi
	s.mu.Unlock()

	if btc == nil || fgi == nil || oi == nil {
		return ComputeResponse{}, errors.New("params service not started")
	}

	snapshot := btc.Snapshot()
	fgiValue := fgi.GetValue()
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
		if record, ok := oi.Get(req.SymbolName); ok {
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
	diag.GainFinal = clampFloat(gainFinal, gainMin, gainMax)
	diag.TwapFinal = clampFloat(twapFinal, twapMin, twapMax)

	resp := ComputeResponse{
		GainPct: diag.GainFinal,
		TwapSec: diag.TwapFinal,
		Diag:    diag,
	}
	return resp, nil
}
