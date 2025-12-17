package toUpbitListBnSymbol

import (
	"time"

	"upbitBnServer/internal/quant/exchanges/binance/bnConst"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/pkg/container/pool/byteBufPool"
	"upbitBnServer/pkg/utils/convertx"

	"github.com/tidwall/gjson"
)

func (s *Single) onBookTick(len int, bufPtr *[]byte) {
	defer byteBufPool.ReleaseBuffer(bufPtr)
	data := (*bufPtr)[:len]

	/****处理盘口数据****/
	if toUpBitListDataAfter.LoadTrig() {
		if s.SymbolIndex != toUpBitListDataAfter.TrigSymbolIndex {
			return
		}
		/*********************上币已经触发**************************/
		results := gjson.GetManyBytes(data, "b", "T")
		bidF64 := results[0].Float()
		// 1、回调执行引擎的函数,判断止盈止损
		s.onBookTickExecute(bidF64, results[1].Int())
	} else {
		/*********************上币还未触发**************************/
		results := gjson.GetManyBytes(data, jsonEvent, "a")
		eventTs := results[0].Int()
		ask_u8 := convertx.PriceStringToUint64(results[1].String(), bnConst.PScale_8)

		// 找出接受到这个markPrice之后的最小ask
		if eventTs > s.markPriceTs {
			if ask_u8 < s.minPriceAfterMp {
				s.minPriceAfterMp = ask_u8
			}
		}
		// 数据太旧则丢弃
		if eventTs <= s.committedTs {
			s.btLatencyTotal.Record(s.StMeta.SymbolName, float64(time.Now().UnixMicro()-1000*eventTs)) // 记录总延迟
			return
		}
		// 2、计算触发信号
		s.checkMarket(eventTs, "bookTick", ask_u8)
		s.btLatencyTotal.Record(s.StMeta.SymbolName, float64(time.Now().UnixMicro()-1000*eventTs))
	}
}
