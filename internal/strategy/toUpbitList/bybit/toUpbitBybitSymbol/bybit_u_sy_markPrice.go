package toUpbitBybitSymbol

import (
	"time"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/pkg/utils/byteUtils"
	"upbitBnServer/pkg/utils/convertx/byteConvert"
)

func (s *Single) CancelPreOrder() {
	if s.pre != nil {
		s.pre.CancelPreOrder(s.symbolName, from_bybit)
		time.Sleep(100 * time.Millisecond) //每秒10次
	} else {
		toUpBitDataStatic.DyLog.GetLog().Errorf("撤单失败,[%d][%s] s.pre is nil", s.symbolIndex, s.symbolName)
	}
}

func getMarketPrice(byteLen, symbolLen uint16, bb []byte) (markPrice float64, ok bool) {
	var loop_begin uint16
	if bb[18+symbolLen+10] == 'd' {
		loop_begin = 18 + symbolLen + 35 + symbolLen + 3
	} else {
		loop_begin = 18 + symbolLen + 38 + symbolLen + 3
	}
	for range 50 {
		if bb[loop_begin] == 'm' && bb[loop_begin+1] == 'a' && bb[loop_begin+2] == 'r' && bb[loop_begin+3] == 'k' {
			// 找到了
			p_begin := loop_begin + 12
			p_end := byteUtils.FindNextQuoteIndex(bb, p_begin, byteLen)
			return byteConvert.ByteArrToF64(bb[p_begin:p_end]), true
		} else {
			loop_begin = byteUtils.FindNextCommaIndex(bb, loop_begin, byteLen) + 2
			if bb[loop_begin] == 'c' {
				// 找完了都不存在
				break
			}
		}
	}
	return 0, false
}

func (s *Single) onMarkPrice(b []byte) {
	if toUpBitListDataAfter.LoadTrig() {
		/*********************上币已经触发**************************/
		if s.symbolIndex != toUpBitListDataAfter.TrigSymbolIndex {
			return
		}
		byteLen := uint16(len(b))
		// 1、计算价格上限并存储
		msE := byteConvert.BytesToInt64(b[byteLen-14 : byteLen-1])
		markPrice, ok := getMarketPrice(byteLen, s.symbolLen, b)
		if !ok {
			return
		}
		priceMaxBuy := markPrice * s.upLimitPercent
		s.trigPriceMax.Store(msE/1000, priceMaxBuy)
		toUpBitDataStatic.DyLog.GetLog().Infof("%s最新[u8:%.8f,u10:%.8f]标记价格: %s", s.symbolName, markPrice, priceMaxBuy, string(b))
	} else {
		markPrice, ok := getMarketPrice(uint16(len(b)), s.symbolLen, b)
		if !ok {
			return
		}
		// 2、计算价格上限
		s.priceMaxBuy = markPrice * s.upLimitPercent
		// 3、回调函数更新预挂单
		s.pre.CheckPreOrder(s.symbolName, markPrice, s.pScale, s.qScale)
	}
}

//{"e":"markPriceUpdate","E":1761731536000,"s":"SOLUSDT","p":"194.36713158","P":"194.24338995","i":"194.44842105","r":"-0.00003843","T":1761753600000}
