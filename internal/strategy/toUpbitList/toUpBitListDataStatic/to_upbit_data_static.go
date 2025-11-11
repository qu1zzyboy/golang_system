package toUpBitListDataStatic

import (
	"strconv"

	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/observe/log/staticLog"
	"upbitBnServer/internal/infra/observe/notify/notifyTg"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/utils/algorithms"
	"upbitBnServer/pkg/container/map/myMap"
	"upbitBnServer/pkg/utils/idGen"
	"upbitBnServer/pkg/utils/timeUtils"

	"github.com/shopspring/decimal"
)

const (
	DAY_BEGIN_STR     = "1 0 7 * * *"   // 每天的 07:00:01 执行任务
	DAY_END_STR       = "1 10 18 * * *" // 每天的 18:10:01 执行任务
	TO_UPBIT_LIST_CFG = "to_upbit_list_cfg"
)

var (
	GlobalCfg         ConfigVir                                                       // 全局配置
	SymbolIndex       = myMap.NewMySyncMap[string, systemx.SymbolIndex16I]()          // symbolName --> symbolIndex
	SymbolMaxNotional = myMap.NewMySyncMap[systemx.SymbolIndex16I, decimal.Decimal]() //symbolIndex-->最大仓位上限
	DyLog             = dynamicLog.NewDynamicLogger(staticLog.Config{                 // 创建日志记录器
		NeedErrorHook: true,
		FileDir:       "toUpBitList",
		DateStr:       timeUtils.GetNowDateStr(),
		FileName:      "instanceId",
		Level:         staticLog.INFO_LEVEL,
	})
	SigLog = dynamicLog.NewDynamicLogger(staticLog.Config{ // 创建日志记录器
		NeedErrorHook: true,
		FileDir:       "toUpBitList",
		DateStr:       timeUtils.GetNowDateStr(),
		FileName:      "signal",
		Level:         staticLog.INFO_LEVEL,
	})
	ExType exchangeEnum.ExchangeType // 交易所类型
	AcType exchangeEnum.AccountType  // 账户类型
)

func getClientOrderId(acType exchangeEnum.AccountType, flag string) string {
	str := ""
	switch acType {
	case exchangeEnum.SPOT:
		str = "sp"
	case exchangeEnum.FUTURE:
		str = "fu"
	case exchangeEnum.SWAP:
		str = "sw"
	case exchangeEnum.FULL_MARGIN:
		str = "fm"
	case exchangeEnum.ISOLATED_MARGIN:
		str = "im"
	default:
	}
	str += strconv.Itoa(algorithms.GetRandom09()) + "-" + string(algorithms.GetRandomaZ()) + "-" + flag
	return str + idGen.GetSnowflakeIdStr()
}

func GetMakerClientOrderId() string {
	return getClientOrderId(AcType, "maker")
}

func GetClientOrderIdBy(flag string) string {
	return getClientOrderId(AcType, flag)
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
