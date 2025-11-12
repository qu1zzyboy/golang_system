package toUpbitBybitSymbol

import (
	"context"
	"time"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitParam"
)

func (s *Single) startTrig() {
	s.ctxStop, s.cancel = context.WithCancel(context.Background())
	var ok bool
	s.maxNotional, ok = toUpBitDataStatic.SymbolMaxNotional.Load(s.symbolIndex)
	if !ok {
		s.maxNotional = 50000
	}
	s.maxNotional = s.maxNotional - toUpbitParam.Dec500
}

func (s *Single) setExecuteParam(trigPrice float64, twapSec float64) {
	s.twapSec = twapSec
	s.takeProfitPrice = trigPrice
	s.closeDuration = time.Duration(twapSec) * time.Second
	toUpBitDataStatic.DyLog.GetLog().Infof("止盈价格: %.8f,平仓持续时间: %s,单账户上限:%s", trigPrice, s.closeDuration.String(), s.maxNotional)
}
