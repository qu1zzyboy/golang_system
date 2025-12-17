package toUpBitListDataAfter

import (
	"sync/atomic"
	"upbitBnServer/internal/strategy/newsDrive/common/driverStatic"
)

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
	driverStatic.DyLog.GetLog().Info("===========================清空ClearTrig()===============================")
	hasTrig.Store(false)
	TrigSymbolIndex = -1
}
