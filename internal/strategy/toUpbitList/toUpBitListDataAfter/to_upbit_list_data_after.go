package toUpBitListDataAfter

import (
	"sync/atomic"

	"github.com/hhh500/quantGoInfra/pkg/container/map/myMap"
	"github.com/hhh500/upbitBnServer/internal/quant/execute"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"github.com/shopspring/decimal"
)

type OnSuccessEvt struct {
	ClientOrderId string
	Volume        decimal.Decimal
	SortPrice     float64
	TimeStamp     int64
	OrderMode     execute.MyOrderMode
	IsOnline      bool
	AccountKeyId  uint8
	InstanceId    orderBelongEnum.Type
}
type OnSuccessOrder func(evt OnSuccessEvt)

var (
	TrigPriceMax_10 = myMap.NewMySyncMap[int64, uint64]()      // 已触发品种的买入上限
	TrigSymbolName  string                                     // 触发的交易对名称,过滤用
	TrigSymbolIndex int                                   = -1 // 触发的交易对索引
	trigBidPrice    atomic.Value                               // 最新买一价格,平仓用,bookTick更新
	hasTrig         atomic.Bool                                // 是否已经触发过价格变化
	HasTreeNews     atomic.Bool                                // 是否已经触发过TreeNews
)

func SaveBidPrice(f64 float64) {
	trigBidPrice.Store(f64)
}

func LoadBidPrice() (float64, bool) {
	val := trigBidPrice.Load()
	if val == nil {
		return 0, false
	}
	return val.(float64), true
}

func Trig(symbolName string, symbolIndex int) {
	trigBidPrice.Store(0.0)
	hasTrig.Store(true)
	TrigSymbolName = symbolName
	TrigSymbolIndex = symbolIndex
	HasTreeNews.Store(false)
}

func LoadTrig() bool {
	return hasTrig.Load()
}

func ClearTrig() {
	toUpBitListDataStatic.DyLog.GetLog().Info("===========================清空ClearTrig()===============================")
	hasTrig.Store(false)
	HasTreeNews.Store(false)
	TrigSymbolName = ""
	TrigSymbolIndex = -1
	TrigPriceMax_10.Clear()
}

func UpdateTreeNewsFlag() {
	HasTreeNews.Store(true)
	toUpBitListDataStatic.SendToUpBitMsg("TreeNews确认", map[string]string{
		"symbol": TrigSymbolName,
		"op":     "TreeNews确认",
	})
}
