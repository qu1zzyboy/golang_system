package toUpbitListBnSymbol

import (
	"time"

	"github.com/hhh500/quantGoInfra/pkg/container/pool/byteBufPool"
	"github.com/hhh500/quantGoInfra/pkg/utils/convertx"
	"github.com/hhh500/quantGoInfra/quant/exchanges/binance/bnConst"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/bn/toUpBitListBnExecute"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"github.com/tidwall/gjson"
)

func (s *Single) onBookTick(len int, bufPtr *[]byte) {
	defer byteBufPool.ReleaseBuffer(bufPtr)
	data := (*bufPtr)[:len]

	/****处理盘口数据****/
	if toUpBitListDataAfter.LoadTrig() {
		if s.symbolIndex != toUpBitListDataAfter.TrigSymbolIndex {
			return
		}
		/*********************上币已经触发**************************/
		results := gjson.GetManyBytes(data, "b", "T")
		bidF64 := results[0].Float()
		toUpBitListDataAfter.SaveBidPrice(bidF64)
		// 1、回调执行引擎的函数,判断止盈止损
		toUpBitListBnExecute.GetExecute().OnBookTick(bidF64, results[1].Int())
	} else {
		/*********************上币还未触发**************************/
		results := gjson.GetManyBytes(data, jsonEvent, "a")
		eventTs := results[0].Int()
		// 数据太旧则丢弃
		if eventTs <= s.committedTs {
			s.btLatencyTotal.Record(s.StMeta.SymbolName, float64(time.Now().UnixMicro()-1000*eventTs)) // 记录总延迟
			return
		}
		// 2、计算触发信号
		s.checkMarket(eventTs, "bookTick", convertx.PriceStringToUint64(results[1].String(), bnConst.PScale_8))
		s.btLatencyTotal.Record(s.StMeta.SymbolName, float64(time.Now().UnixMicro()-1000*eventTs))
	}
}
