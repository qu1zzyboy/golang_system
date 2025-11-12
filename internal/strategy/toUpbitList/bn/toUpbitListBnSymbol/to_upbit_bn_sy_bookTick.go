package toUpbitListBnSymbol

import (
	"time"

	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/pkg/utils/byteUtils"
	"upbitBnServer/pkg/utils/convertx/byteConvert"
)

func (s *Single) onBookTick(byteLen uint16, b []byte) {
	/****处理盘口数据****/
	if toUpBitListDataAfter.LoadTrig() {
		if s.symbolIndex != toUpBitListDataAfter.TrigSymbolIndex {
			return
		}
		/*********************上币已经触发**************************/
		msT := byteConvert.BytesToInt64(b[byteLen-32 : byteLen-19])
		b_start := 48 + s.symbolLen
		b_end := byteUtils.FindNextQuoteIndex(b, b_start, byteLen)
		bid := byteConvert.ByteArrToF64(b[b_start:b_end])
		// 1、回调执行引擎的函数,判断止盈止损
		s.onBookTickExecute(bid, msT)
	} else {
		/*********************上币还未触发**************************/
		msT := byteConvert.BytesToInt64(b[byteLen-32 : byteLen-19])
		msE := byteConvert.BytesToInt64(b[byteLen-14 : byteLen-1])
		b_start := 48 + s.symbolLen
		b_end := byteUtils.FindNextQuoteIndex(b, b_start, byteLen)
		bid := byteConvert.ByteArrToF64(b[b_start:b_end])

		// 找出接受到这个markPrice之后的最小ask
		if msT > s.markPriceTs {
			if bid < s.minPriceAfterMp {
				s.minPriceAfterMp = bid
			}
		}
		// 2、计算触发信号
		s.checkMarket(msT, bid)
		s.btLatencyTotal.Record(s.symbolName, float64(time.Now().UnixMicro()-1000*msE))
	}
}
