package toUpbitListBnSymbol

import (
	"upbitBnServer/internal/infra/systemx/instanceEnum"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitBnMode"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/pkg/utils/byteUtils"
	"upbitBnServer/pkg/utils/convertx/byteConvert"
)

func (s *Single) CancelPreOrder() {
	if s.pre != nil {
		s.pre.CancelPreOrder(s.can, instanceEnum.TO_UPBIT_LIST_BN)
	} else {
		toUpBitDataStatic.DyLog.GetLog().Errorf("撤单失败,[%d][%s] s.pre is nil", s.symbolIndex, s.symbolName)
	}
}

func (s *Single) onMarkPrice(byteLen uint16, b []byte) {
	if toUpBitListDataAfter.LoadTrig() {
		/*********************上币已经触发**************************/
		if s.symbolIndex != toUpBitListDataAfter.TrigSymbolIndex {
			return
		}
		// 1、计算价格上限并存储
		msE := byteConvert.BytesToInt64(b[27:40])
		p_start := 46 + s.symbolLen + 7
		p_end := byteUtils.FindNextQuoteIndex(b, p_start, byteLen)
		s.thisMarkPrice = byteConvert.ByteArrToF64(b[p_start:p_end])
		priceMaxBuy := s.thisMarkPrice * s.upLimitPercent
		s.trigPriceMax.Store(msE/1000, priceMaxBuy)
		toUpBitDataStatic.DyLog.GetLog().Infof("%s最新[u8:%.8f,u10:%.8f]标记价格: %s", s.symbolName, s.thisMarkPrice, priceMaxBuy, string(b))
	} else {
		msE := byteConvert.BytesToInt64(b[27:40])
		p_start := 46 + s.symbolLen + 7
		p_end := byteUtils.FindNextQuoteIndex(b, p_start, byteLen)
		markPrice := byteConvert.ByteArrToF64(b[p_start:p_end])
		// 2、计算价格上限
		s.thisMarkPrice = markPrice
		s.priceMaxBuy = s.thisMarkPrice * s.upLimitPercent
		s.markPriceTs = msE
		s.minPriceAfterMp = markPrice
		// 2、计算价格上限
		if !toUpbitBnMode.Mode.IsPlacePreOrder() {
			return
		}
		// 3、回调函数更新预挂单
		s.pre.CheckPreOrder(s.symbolName, s.thisMarkPrice, s.pScale, s.qScale)
	}
}
