package toUpBitListDataAfter

import (
	"sync/atomic"

	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"

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
	TrigSymbolIndex int         = -1 // 触发的交易对索引
	hasTrig         atomic.Bool      // 是否已经成功触发,这里必须是全局变量,减少cpu解析
)

func Trig(symbolIndex int) {
	hasTrig.Store(true)
	TrigSymbolIndex = symbolIndex
}

func LoadTrig() bool {
	return hasTrig.Load()
}

func ClearTrig() {
	toUpBitDataStatic.DyLog.GetLog().Info("===========================清空ClearTrig()===============================")
	hasTrig.Store(false)
	TrigSymbolIndex = -1
}
