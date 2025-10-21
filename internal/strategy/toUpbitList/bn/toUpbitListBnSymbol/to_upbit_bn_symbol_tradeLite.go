package toUpbitListBnSymbol

import (
	"github.com/hhh500/quantGoInfra/pkg/utils/convertx"
	"github.com/hhh500/quantGoInfra/quant/exchanges/binance/bnConst"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"github.com/tidwall/gjson"
)

func (s *Single) onTradeLite(data []byte) {
	if toUpBitListDataAfter.LoadTrig() {
		if s.symbolIndex == toUpBitListDataAfter.TrigSymbolIndex {
			toUpBitListDataStatic.SendToUpBitMsg("发送bn二次确认失败", map[string]string{
				"symbol": toUpBitListDataAfter.TrigSymbolName,
				"op":     "bn_TRADE_LITE二次上涨确认",
			})
			toUpBitListDataStatic.DyLog.GetLog().Infof("触发后二次确认:%s", toUpBitListDataAfter.TrigSymbolName)
		}
	} else {
		/*********************上币还未触发**************************/
		eventTs := gjson.GetBytes(data, jsonEvent).Int()
		go func() {
			toUpBitListDataStatic.DyLog.GetLog().Infof("[%d]触发前成交,%s", s.accountKeyId, string(data))
			toUpBitListDataStatic.SendToUpBitMsg("发送bn预挂单成交失败", map[string]string{
				"symbol": s.StMeta.SymbolName,
				"op":     "bn_TRADE_LITE预挂单成交",
			})
		}()
		riseValue := 0.0
		var priceU64_8 uint64 = 0
		if eventTs <= s.committedTs {
			riseValue = s.lastRiseValue
		} else {
			priceU64_8 = convertx.PriceStringToUint64(gjson.GetBytes(data, "p").String(), bnConst.PScale_8)
			riseValue, _, _ = s.commit(priceU64_8, toUpBitListDataStatic.OrderRiceTrig, eventTs)
		}
		s.onOrderPriceCheck(eventTs, priceU64_8, riseValue)
		s.IntoExecuteCheck(eventTs, "preOrder", riseValue, priceU64_8)
	}
	s.onPreFilled(gjson.GetBytes(data, "c").String())
}
