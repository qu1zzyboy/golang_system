package driverParam

import (
	"context"
	"time"
	"upbitBnServer/internal/define/defineTime"
	"upbitBnServer/internal/infra/global/globalCron"
	"upbitBnServer/internal/strategy/newsDrive/common/driverStatic"

	"upbitBnServer/pkg/container/map/myMap"

	"github.com/go-redis/redis/v8"
	"github.com/tidwall/gjson"
)

const bn_open_interest = "BN_OPEN_INTEREST"

// OIConfig 描述 OI 快照的刷新策略。
type OIConfig struct {
	RefreshEvery time.Duration
	Timeout      time.Duration
}

// OIRecord 为按交易对整理后的结构化数据。
type OIRecord struct {
	OpenInterest float64
	OpenQty      float64
	Timestamp    int64
}

type OIStore struct {
	cfg      OIConfig
	data     myMap.MySyncMap[uint64, *OIRecord]
	redisCfg *redis.Client
}

func (p *OIStore) LoadBySymbolIndex(symbolIndex int) (*OIRecord, bool) {
	return p.data.Load(uint64(symbolIndex))
}

func NewOIStore(cfg OIConfig, redisCfg *redis.Client) *OIStore {
	if cfg.RefreshEvery <= 0 {
		cfg.RefreshEvery = 60 * time.Second
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 1500 * time.Millisecond
	}
	return &OIStore{
		cfg:      cfg,
		data:     myMap.NewMySyncMap[uint64, *OIRecord](),
		redisCfg: redisCfg,
	}
}

func (p *OIStore) Start() error {
	_, err := globalCron.AddFunc(defineTime.MinEndStr_59, func() {
		if err := p.fetchOnce(); err != nil {
			logError.GetLog().Errorf("刷新OI信息出错,%v", err)
		}
	})
	if err != nil {
		return err
	}
	return nil
}

func (p *OIStore) fetchOnce() error {
	res := p.redisCfg.HGetAll(context.Background(), bn_open_interest)
	if res.Err() != nil {
		return res.Err()
	}
	data, err := res.Result()
	if err != nil {
		return err
	}
	for _, jsonStr := range data {
		symbolName := gjson.Get(jsonStr, "symbol").String()
		symbolIndex, ok := driverStatic.SymbolIndex.Load(symbolName)
		if !ok {
			continue
		}
		var temp OIRecord
		temp.Timestamp = gjson.Get(jsonStr, "time").Int()
		temp.OpenInterest = gjson.Get(jsonStr, "open_interest").Float()
		temp.OpenQty = gjson.Get(jsonStr, "open_qty").Float()
		p.data.Store(uint64(symbolIndex), &temp)
	}
	return nil
}

// {
//   "symbol": "1000RATSUSDT",
//   "time_str": "2025-10-24 14:29:32.119",
//   "open_interest": 200626085,
//   "open_qty": 5788062.55225,
//   "time": 1761287372119
// }
