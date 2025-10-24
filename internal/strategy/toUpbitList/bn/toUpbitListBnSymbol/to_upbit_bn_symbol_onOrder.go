package toUpbitListBnSymbol

import (
	"upbitBnServer/internal/quant/exchanges/binance/bnConst"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"upbitBnServer/pkg/utils/convertx"

	"github.com/tidwall/gjson"
)

func (s *Single) onPayloadOrder(data []byte) {
	if toUpBitListDataAfter.LoadTrig() {
		if s.symbolIndex == toUpBitListDataAfter.TrigSymbolIndex {
			toUpBitListDataStatic.SendToUpBitMsg("发送bn二次确认失败", map[string]string{
				"symbol": s.StMeta.SymbolName,
				"op":     "ORDER_UPDATE二次上涨确认",
			})
			toUpBitListDataStatic.DyLog.GetLog().Infof("触发后二次确认:%s", s.StMeta.SymbolName)
		}
	} else {
		/*********************上币还未触发**************************/
		eventTs := gjson.GetBytes(data, jsonEvent).Int()
		go func() {
			toUpBitListDataStatic.DyLog.GetLog().Infof("[%d]触发前成交,%s", s.preAccountKeyId, string(data))
			toUpBitListDataStatic.SendToUpBitMsg("发送bn预挂单成交失败", map[string]string{
				"symbol": s.StMeta.SymbolName,
				"op":     "ORDER_UPDATE预挂单成交",
			})
		}()
		priceU64_8 := convertx.PriceStringToUint64(gjson.GetBytes(data, "o.p").String(), bnConst.PScale_8)
		s.onOrderPriceCheck(eventTs, priceU64_8)
	}
	s.onPreFilled(gjson.GetBytes(data, "o.c").String())
}
