package driverParam

import (
	"context"
	"time"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/quant/market/kline/klineBnRest"
	"upbitBnServer/internal/quant/market/kline/klineEnum"
	"upbitBnServer/internal/quant/market/kline/klineModel"
	"upbitBnServer/internal/quant/market/kline/klineSubBn"
	"upbitBnServer/pkg/container/map/myMap"
	"upbitBnServer/pkg/container/pool/byteBufPool"
	"upbitBnServer/pkg/utils/timeUtils"

	"github.com/tidwall/gjson"
)

const symbol = "BTCUSDT"

// BTCConfig 对应 Python 版本的可调参数。
type BTCConfig struct {
	Symbol            string
	H1WindowSize      int
	H1RegularPullSec  int
	H1EdgePreSec      int
	H1EdgePostSec     int
	M1PullSec         int
	StartReadyTimeout time.Duration
}

// BTCMetrics 周期性获取币安合约 K 线，维护 24h/7d 收益计算所需的历史数据，并保证并发安全。
type BTCMetrics struct {
	cfg  BTCConfig
	data myMap.MySyncMap[int64, float64] // 1h级别k线的开盘时间戳-->收盘价
}

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
	if cfg.StartReadyTimeout <= 0 {
		cfg.StartReadyTimeout = 20 * time.Second
	}
	return &BTCMetrics{
		cfg:  cfg,
		data: myMap.NewMySyncMap[int64, float64](),
	}
}

func (m *BTCMetrics) Start(ctx context.Context) error {
	resp, err := klineBnRest.BnImpl.GetKlineSlice(&klineModel.KlineReq{
		SymbolName: symbol,
		KlineSize:  144,
		Interval:   klineEnum.KLINE_1h,
		AcType:     exchangeEnum.FUTURE,
	})
	if err != nil {
		return err
	}
	for _, v := range resp {
		m.data.Store(v.OpenTimeStamp, v.ClosePrice)
	}
	if err := klineSubBn.GetManager().RegisterReadHandler(ctx, []string{symbol}, m.OnKLine); err != nil {
		return err
	}
	return nil
}

func (m *BTCMetrics) Snapshot(nowMs int64) BTCSnapshot {
	hourTsBegin := timeUtils.ConvertMillTs2HourStartMill(nowMs)
	this_close, ok := m.data.Load(hourTsBegin)
	if !ok {
		logError.GetLog().Errorf("当前BTCUSDT 1h 收盘价获取失败,时间戳[%d,%d]", hourTsBegin, nowMs)
		return BTCSnapshot{
			BTC1D: 0,
			BTC7D: 0,
		}
	}
	last_24h_close, ok := m.data.Load(hourTsBegin - 23*60*60*1000)
	if !ok {
		logError.GetLog().Errorf("当前BTCUSDT 24h 收盘价获取失败,时间戳[%d,%d]", hourTsBegin-23*60*60*1000, nowMs)
		return BTCSnapshot{
			BTC1D: 0,
			BTC7D: 0,
		}
	}
	last_144h_close, ok := m.data.Load(hourTsBegin - 143*60*60*1000)
	if !ok {
		logError.GetLog().Errorf("当前BTCUSDT 144h 收盘价获取失败,时间戳[%d,%d]", hourTsBegin-143*60*60*1000, nowMs)
		return BTCSnapshot{
			BTC1D: 0,
			BTC7D: 0,
		}
	}
	return BTCSnapshot{
		BTC1D: 100 * (this_close - last_24h_close) / last_24h_close,
		BTC7D: 100 * (this_close - last_144h_close) / last_144h_close,
	}
}

func (m *BTCMetrics) OnKLine(len int, bufPtr *[]byte) {
	data := (*bufPtr)[:len]
	defer byteBufPool.ReleaseBuffer(bufPtr)
	openTs := gjson.GetBytes(data, "k.t").Int()
	closePrice := gjson.GetBytes(data, "k.c").Float()
	m.data.Store(openTs, closePrice)

	// 新小时推送,有145根k线
	if m.data.Length() > 144 {
		m.data.Delete(openTs - 144*60*60*1000)
	}
}

func (m *BTCMetrics) Stop(ctx context.Context) error {
	return nil
}
