package toUpbitListBnSymbol

import (
	"context"
	"time"

	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"

	"github.com/shopspring/decimal"
)

func (s *Single) startTrig() {
	s.ctxStop, s.cancel = context.WithCancel(context.Background())
	var ok bool
	s.maxNotional, ok = toUpBitListDataStatic.SymbolMaxNotional.Load(s.symbolIndex)
	if !ok {
		s.maxNotional = decimal.NewFromInt(50000)
	}
	s.maxNotional = s.maxNotional.Sub(toUpBitListDataStatic.Dec500)
}

func (s *Single) setExecuteParam(trigPrice float64, twapSec float64) {
	s.twapSec = twapSec
	s.takeProfitPrice = trigPrice
	s.closeDuration = time.Duration(twapSec) * time.Second
	toUpBitListDataStatic.DyLog.GetLog().Infof("止盈价格: %.8f,平仓持续时间: %s,单账户上限:%s", trigPrice, s.closeDuration.String(), s.maxNotional)
}
