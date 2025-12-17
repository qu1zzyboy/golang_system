package bnDriveSymbol

import (
	"upbitBnServer/internal/quant/exchanges/binance/bnConst"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/pkg/container/pool/byteBufPool"
	"upbitBnServer/pkg/utils/convertx"

	"github.com/tidwall/gjson"
)

func (s *Single) onAggTrade(len int, bufPtr *[]byte) {
	defer byteBufPool.ReleaseBuffer(bufPtr)
	data := (*bufPtr)[:len]

	/****处理成交数据****/
	if toUpBitListDataAfter.LoadTrig() {
		/*********************上币已经触发**************************/
		if s.symbolIndex != toUpBitListDataAfter.TrigSymbolIndex {
			return
		}
	} else {
		/*********************上币还未触发**************************/
		eventTs := gjson.GetBytes(data, jsonEvent).Int()
		// 数据太旧则丢弃
		if eventTs <= s.committedTs {
			return
		}
		// 2、计算触发信号
		s.checkMarket(eventTs, "aggTrade", convertx.PriceStringToUint64(gjson.GetBytes(data, "p").String(), bnConst.PScale_8))
	}
}
