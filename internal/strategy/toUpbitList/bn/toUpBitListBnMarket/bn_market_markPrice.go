package toUpBitListBnMarket

import (
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListChan"
	"upbitBnServer/pkg/container/pool/byteBufPool"

	"github.com/tidwall/gjson"
)

func (s *Market) OnMarkPricePool(len int, bufPtr *[]byte) {
	data := (*bufPtr)[:len]
	result := gjson.GetBytes(data, jsonSymbol)
	if !result.Exists() {
		if !gjson.GetBytes(data, "id").Exists() {
			toUpBitDataStatic.DyLog.GetLog().Errorf("markPrice symbol not found: %s", string(data))
		}
		byteBufPool.ReleaseBuffer(bufPtr)
		return
	}
	symbolIndex, ok := toUpBitDataStatic.SymbolIndex.Load(result.String())
	if !ok {
		toUpBitDataStatic.DyLog.GetLog().Errorf("markPrice symbol not found: %s", string(data))
		byteBufPool.ReleaseBuffer(bufPtr)
		return
	}
	toUpbitListChan.SendMarkPrice(symbolIndex, bufPtr, len)
}

//

// {
// 	"e": "markPriceUpdate",
// 	"E": 1760239703000,
// 	"s": "TAUSDT",
// 	"p": "0.04080855",
// 	"P": "0.04031165",
// 	"i": "0.04072079",
// 	"r": "0.00005000",
// 	"T": 1760241600000
// }
