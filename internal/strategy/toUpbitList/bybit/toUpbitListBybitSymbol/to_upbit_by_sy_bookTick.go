package toUpbitListBybitSymbol

import (
	"time"

	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/pkg/utils/byteUtils"
	"upbitBnServer/pkg/utils/convertx"
)

func (s *Single) onBookTick(byteLen uint16, b []byte) {
	/****处理盘口数据****/
	if toUpBitListDataAfter.LoadTrig() {
		if s.symbolIndex != toUpBitListDataAfter.TrigSymbolIndex {
			return
		}
		/*********************上币已经触发**************************/
		msT := convertx.BytesToInt64(b[byteLen-14 : byteLen-1])
		b_start := 22 + s.symbolLen + 25 + 13 + 14 + s.symbolLen + 9
		b_end := byteUtils.FindNextQuoteIndex(b, b_start, byteLen)
		bid_u8 := convertx.PriceByteArrToUint64(b[b_start:b_end], 8)
		// 1、回调执行引擎的函数,判断止盈止损
		s.onBookTickExecute(float64(bid_u8)/1e8, msT)
	} else {
		/*********************上币还未触发**************************/
		msT := convertx.BytesToInt64(b[byteLen-14 : byteLen-1])

		var e_begin = 22 + s.symbolLen + 25
		msE := convertx.BytesToInt64(b[e_begin : e_begin+13])
		b_start := e_begin + 13 + 14 + s.symbolLen + 9
		b_end := byteUtils.FindNextQuoteIndex(b, b_start, byteLen)
		bid_u8 := convertx.PriceByteArrToUint64(b[b_start:b_end], 8)
		// 数据太旧则丢弃
		if msT <= s.committedTs {
			s.btLatencyTotal.Record(s.StMeta.SymbolName, float64(time.Now().UnixMicro()-1000*msE)) // 记录总延迟
			return
		}
		// 2、计算触发信号
		s.checkMarket(msT, bid_u8)
		s.btLatencyTotal.Record(s.StMeta.SymbolName, float64(time.Now().UnixMicro()-1000*msE))
	}
}
