package toUpBitListBnMarket

import (
	"upbitBnServer/internal/strategy/newsDrive/common/driverListChan"
	"upbitBnServer/internal/strategy/newsDrive/common/driverStatic"
	"upbitBnServer/pkg/container/pool/byteBufPool"

	"github.com/tidwall/gjson"
)

func (s *Market) OnAggTradePool(len int, bufPtr *[]byte) {
	data := (*bufPtr)[:len]
	result := gjson.GetBytes(data, jsonSymbol)
	if !result.Exists() {
		if !gjson.GetBytes(data, "id").Exists() {
			driverStatic.DyLog.GetLog().Errorf("aggTrade symbol not found: %s", string(data))
		}
		byteBufPool.ReleaseBuffer(bufPtr)
		return
	}
	symbolIndex, ok := driverStatic.SymbolIndex.Load(result.String())
	if !ok {
		driverStatic.DyLog.GetLog().Errorf("aggTrade symbol not found: %s", string(data))
		byteBufPool.ReleaseBuffer(bufPtr)
		return
	}
	driverListChan.SendAggTrade(symbolIndex, bufPtr, len)
}

// {
// 	"e": "aggTrade",
// 	"E": 1749721515950,
// 	"a": 125390584,
// 	"s": "ETHUSDC",
// 	"p": "2755.00",
// 	"q": "0.181",
// 	"f": 221147828,
// 	"l": 221147828,
// 	"T": 1749721515794, //第一笔归集的时间戳
// 	"m": true
// }
