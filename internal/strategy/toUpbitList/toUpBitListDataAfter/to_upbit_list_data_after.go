package toUpBitListDataAfter

import (
	"sync/atomic"

	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
)

type OnSuccessEvt struct {
	ClientOrderId systemx.WsId16B
	Volume        float64
	TimeStamp     int64
	OrderMode     execute.OrderMode
	IsOnline      bool
	AccountKeyId  uint8
}

type OnSuccessOrder func(evt OnSuccessEvt)

const TrigIndexDefault systemx.SymbolIndex16I = -1

var (
	TrigSymbolIndex systemx.SymbolIndex16I = TrigIndexDefault // 触发5%的交易对索引
	hasTrig         atomic.Bool                               // 是否已经成功触发,这里必须是全局变量,减少cpu解析
)

func Trig(symbolIndex systemx.SymbolIndex16I) {
	hasTrig.Store(true)
	TrigSymbolIndex = symbolIndex
}

func LoadTrig() bool {
	return hasTrig.Load()
}

func ClearTrig() {
	toUpBitDataStatic.DyLog.GetLog().Info("===========================清空ClearTrig()===============================")
	hasTrig.Store(false)
	TrigSymbolIndex = TrigIndexDefault
}
