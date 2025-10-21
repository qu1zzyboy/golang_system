package toUpBitListBnMarket

import (
	"github.com/hhh500/quantGoInfra/pkg/container/pool/byteBufPool"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpbitListChan"
	"github.com/tidwall/gjson"
)

func (s *Market) OnBookTickPool(len int, bufPtr *[]byte) {
	data := (*bufPtr)[:len]
	result := gjson.GetBytes(data, jsonSymbol)
	if !result.Exists() {
		if !gjson.GetBytes(data, "id").Exists() {
			toUpBitListDataStatic.DyLog.GetLog().Errorf("bookTick symbol not found: %s", string(data))
		}
		byteBufPool.ReleaseBuffer(bufPtr)
		return
	}
	symbolIndex, ok := toUpBitListDataStatic.SymbolIndex.Load(result.String())
	if !ok {
		toUpBitListDataStatic.DyLog.GetLog().Errorf("bookTick symbol not found: %s", string(data))
		byteBufPool.ReleaseBuffer(bufPtr)
		return
	}
	toUpbitListChan.SendBookTick(symbolIndex, bufPtr, len)
}

//  {
// 		"e": "bookTicker",
// 		"u": 8017868938289,
// 		"s": "TRXUSDT",
// 		"b": "0.30087",
// 		"B": "77473",
// 		"a": "0.30088",
// 		"A": "4653",
// 		"T": 1752462352794,
// 		"E": 1752462352794
// 	}
