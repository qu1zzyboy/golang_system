package toUpbitListBnSymbol

import (
	"upbitBnServer/internal/quant/exchanges/binance/bnConst"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/pkg/utils/convertx"

	"github.com/tidwall/gjson"
)

func (s *Single) onTradeLite(data []byte) {
	if toUpBitListDataAfter.LoadTrig() {
		if s.SymbolIndex == toUpBitListDataAfter.TrigSymbolIndex {
			toUpBitDataStatic.SendToUpBitMsg("发送bn二次确认失败", map[string]string{
				"symbol": s.StMeta.SymbolName,
				"op":     "bn_TRADE_LITE二次上涨确认",
			})
			toUpBitDataStatic.DyLog.GetLog().Infof("触发后二次确认:%s", s.StMeta.SymbolName)
		}
	} else {
		/*********************上币还未触发**************************/
		tradeTs := gjson.GetBytes(data, jsonT).Int()
		go func() {
			toUpBitDataStatic.SigLog.GetLog().Infof("[%d]触发前成交,%s", s.preAccountKeyId, string(data))
			toUpBitDataStatic.SendToUpBitMsg("发送bn预挂单成交失败", map[string]string{
				"symbol": s.StMeta.SymbolName,
				"op":     "bn_TRADE_LITE预挂单成交",
			})
		}()
		priceU64_8 := convertx.PriceStringToUint64(gjson.GetBytes(data, "p").String(), bnConst.PScale_8)
		s.onOrderPriceCheck(tradeTs, priceU64_8)
	}
	s.onPreFilled(gjson.GetBytes(data, "c").String())
}
