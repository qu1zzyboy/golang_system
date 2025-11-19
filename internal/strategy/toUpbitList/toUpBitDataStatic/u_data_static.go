package toUpBitDataStatic

import (
	"upbitBnServer/internal/infra/observe/log/logCfg"
	"upbitBnServer/internal/infra/systemx"

	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/observe/log/staticLog"
	"upbitBnServer/internal/infra/observe/notify/notifyTg"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/pkg/container/map/myMap"
	"upbitBnServer/pkg/container/ring/ringBuf"
	"upbitBnServer/pkg/utils/timeUtils"

	"github.com/shopspring/decimal"
)

const (
	DAY_BEGIN_STR     = "1 0 7 * * *"   // 每天的 07:00:01 执行任务
	DAY_END_STR       = "1 10 18 * * *" // 每天的 18:10:01 执行任务
	TO_UPBIT_LIST_CFG = "to_upbit_list_cfg"
)

var (
	GlobalCfg         ConfigVir                                               // 全局配置
	SymbolMaxNotional = myMap.NewMySyncMap[systemx.SymbolIndex16I, float64]() //symbolIndex-->最大仓位上限
	Dec500            = decimal.NewFromInt(500)                               // 小于这个数全部平仓
	PriceRiceTrig     float64                                                 // 价格触发阈值,当价格变化超过该值时触发
	DyLog             = dynamicLog.NewDynamicLogger(staticLog.Config{         // 创建日志记录器
		NeedErrorHook: true,
		FileDir:       "toUpBitList",
		DateStr:       timeUtils.GetNowDateStr(),
		FileName:      "instanceId",
		Level:         logCfg.G_LOG_LEVEL,
	})
	SigLog = dynamicLog.NewDynamicLogger(staticLog.Config{ // 创建日志记录器
		NeedErrorHook: true,
		FileDir:       "toUpBitList",
		DateStr:       timeUtils.GetNowDateStr(),
		FileName:      "signal",
		Level:         logCfg.G_LOG_LEVEL,
	})
	TickCap ringBuf.Capacity          // 容量
	ExType  exchangeEnum.ExchangeType // 交易所类型
	AcType  exchangeEnum.AccountType  // 账户类型
)

func SetParam(priceRiceTrig float64, tickCap ringBuf.Capacity, dec500 int64) {
	TickCap = tickCap
	PriceRiceTrig = priceRiceTrig
	Dec500 = decimal.NewFromInt(dec500)
}

func UpdateParam(priceRiceTrig float64) {
	PriceRiceTrig = priceRiceTrig
}

func SendToUpBitMsg(flag string, payload map[string]string) {
	go func() {
		if err := notifyTg.GetTg().SendToUpBitMsg(payload); err != nil {
			DyLog.GetLog().Errorf("%s:%v", flag, err)
		}
	}()
}

func SendToUpBitStrMsg(flag, msg string) {
	go func() {
		if err := notifyTg.GetTg().SendToUpBitStrMsg(msg); err != nil {
			DyLog.GetLog().Errorf("%s:%v", flag, err)
		}
	}()
}
