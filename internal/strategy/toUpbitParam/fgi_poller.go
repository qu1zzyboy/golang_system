package toUpbitParam

import (
	"sync/atomic"
	"time"
	"upbitBnServer/internal/define/defineTime"
	"upbitBnServer/internal/infra/global/globalCron"

	"github.com/go-resty/resty/v2"
	"github.com/jpillora/backoff"
	"github.com/tidwall/gjson"
)

var (
	backOff = &backoff.Backoff{
		Min:    2 * time.Second,  //最小重连间隔
		Max:    20 * time.Second, //最大重连间隔
		Factor: 1.8,
		Jitter: false,
	}
)

// FGIConfig 描述恐惧贪婪指数轮询的行为参数。
type FGIConfig struct {
	Interval         time.Duration
	StartReadyWait   time.Duration
	DefaultFGIValue  float64
	FallbackFGIValue float64
}

// FGIPoller 周期性拉取 alternative.me 的指数，并缓存最新结果。
type FGIPoller struct {
	cfg             FGIConfig
	Classification  string       //fgi_描述
	Value           atomic.Value //fgi_value
	Timestamp       int64        //更新时间戳
	TimeUntilUpdate int64
	client          *resty.Client
}

func NewFGIPoller(cfg FGIConfig) *FGIPoller {
	if cfg.Interval <= 0 {
		cfg.Interval = 30 * time.Second
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
	p := &FGIPoller{
		cfg:    cfg,
		client: resty.New().SetHeader("Accept", "application/json"),
	}
	p.Value.Store(cfg.DefaultFGIValue)
	return p
}

func (p *FGIPoller) LoadValue() (float64, bool) {
	if p.Value.Load() == nil {
		return 0, false
	}
	return p.Value.Load().(float64), true
}

func (p *FGIPoller) Start() error {
	p.fetchLoop()
	_, err := globalCron.AddFunc(defineTime.DayBeginStr, func() {
		p.fetchLoop()
	})
	if err != nil {
		return err
	}
	return nil
}

func (p *FGIPoller) fetchLoop() {
	for {
		if err := p.fetchOnce(); err != nil {
			delay := backOff.Duration()
			logError.GetLog().Errorf("[fgi_get]错误: %v,等待时间: %s", err, delay)
			time.Sleep(delay)
			continue
		} else {
			backOff.Reset()
			break
		}
	}
}

func (p *FGIPoller) fetchOnce() error {
	httpRes, err := p.client.R().Get("https://api.alternative.me/fng/?limit=2")
	if err != nil {
		return err
	}
	data := httpRes.Body()
	p.TimeUntilUpdate = gjson.GetBytes(data, "data.0.time_until_update").Int()
	p.Value.Store(gjson.GetBytes(data, "data.1.value").Float())
	p.Classification = gjson.GetBytes(data, "data.1.value_classification").String()
	p.Timestamp = gjson.GetBytes(data, "data.1.timestamp").Int()
	return nil
}

// {
//   "name": "Fear and Greed Index",
//   "data": [
//     {
//       "value": "30",
//       "value_classification": "Fear",
//       "timestamp": "1761264000",
//       "time_until_update": "65920"
//     },
//     {
//       "value": "27",
//       "value_classification": "Fear",
//       "timestamp": "1761177600"
//     }
//   ],
//   "metadata": {
//     "error": null
//   }
// }
