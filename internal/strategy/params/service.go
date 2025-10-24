package params

import (
	"context"
	"math"
	"time"

	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/pkg/singleton"

	"github.com/go-redis/redis/v8"
)

const (
	// ModuleId 用于在 bootx 中注册该模块。
	ModuleId   = "params_service"
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

func GetService() *Service {
	return serviceSingleton.Get()
}

// Service 为参数计算核心，负责整合 BTC 收益、恐惧贪婪指数与币安 OI，并提供同步计算接口。
type Service struct {
	cfg       Config
	factories providerFactories
	btc       BTCProvider
	fgi       *FGIPoller
	oi        *OIStore
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
	}
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
			StartReadyWait:   10 * time.Second,
			DefaultFGIValue:  defaultFGI,
			FallbackFGIValue: defaultFGI,
		},
		OI: OIConfig{
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
	}
}

// SetConfig 在启动前调整配置。
func (s *Service) SetConfig(cfg Config) error {
	s.cfg = cfg
	return nil
}

// SetProviderFactories 在启动前替换默认工厂，方便注入 mock。
func (s *Service) SetProviderFactories(btc func(BTCConfig) BTCProvider, fgi func(FGIConfig) FGIProvider, oi func(OIConfig) OIProvider) error {
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
	s.btc = btc
	return nil
}

// Start 启动所有后台任务。
func (s *Service) Start(ctx context.Context, redisClient *redis.Client) error {
	s.btc = s.factories.newBTC(s.cfg.BTC)
	if err := s.btc.Start(ctx); err != nil {
		return err
	}

	s.fgi = NewFGIPoller(s.cfg.FGI)
	if err := s.fgi.Start(); err != nil {
		return err
	}
	s.oi = NewOIStore(s.cfg.OI, redisClient)
	if err := s.oi.Start(); err != nil {
		return err
	}
	return nil
}

// Stop 优雅关闭后台协程。
func (s *Service) Stop(ctx context.Context) error {
	return nil
}

// Diagnostics 记录观测字段，与 Python 版本保持一致便于对齐。
type Diagnostics struct {
	OI               float64
	OITimestamp      int64
	S                float64
	SNorm            float64
	GainBase         float64
	GainOIAdd        float64
	GainFinal        float64
	TwapBase         float64
	TwapOIAdd        float64
	TwapFinal        float64
	FGI              float64
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
func (s *Service) Compute(ctx context.Context, isMeme bool, symbolIndex int, marketCapM float64) (ComputeResponse, error) {
	fgiValue, ok := s.fgi.LoadValue()
	if !ok || fgiValue <= 0 {
		fgiValue = defaultFGI
	}
	snapshot := s.btc.Snapshot()

	btc1d := snapshot.BTC1D
	if math.IsNaN(btc1d) {
		btc1d = 0
	}
	btc7d := snapshot.BTC7D
	if math.IsNaN(btc7d) {
		btc7d = 0
	}
	gainBase, twapBase := expectedSplitGainAndTwapDuration(marketCapM, fgiValue, btc1d, btc7d, isMeme)

	var (
		oiValue float64
		oiTime  int64
	)
	if record, ok := s.oi.LoadBySymbolIndex(symbolIndex); ok {
		oiValue = record.OpenInterest
		if record.Timestamp > 0 {
			oiTime = record.Timestamp
		}
	}

	gainAdd, twapAdd, strength, norm := computeOIContribs(oiValue, marketCapM)
	gainFinal, gainMin, gainMax := clipGain(marketCapM, gainBase+gainAdd)
	twapFinal, twapMin, twapMax := clipTwap(marketCapM, twapBase+twapAdd)

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
	diag.OI = oiValue
	diag.OITimestamp = oiTime
	diag.S = strength
	diag.SNorm = norm
	diag.GainFinal = clampFloat(gainFinal, gainMin, gainMax)
	diag.TwapFinal = clampFloat(twapFinal, twapMin, twapMax)

	resp := ComputeResponse{
		GainPct: diag.GainFinal,
		TwapSec: diag.TwapFinal,
		Diag:    diag,
	}
	return resp, nil
}
