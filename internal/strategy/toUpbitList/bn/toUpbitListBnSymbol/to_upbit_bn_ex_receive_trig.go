package toUpbitListBnSymbol

import (
	"context"
	"time"

	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"

	"github.com/shopspring/decimal"
)

func (s *Single) startTrig() {
	s.ctxStop, s.cancel = context.WithCancel(context.Background())
	var ok bool
	s.MaxNotional, ok = toUpBitDataStatic.SymbolMaxNotional.Load(s.SymbolIndex)
	if !ok {
		s.MaxNotional = decimal.NewFromInt(50000)
	}
	s.MaxNotional = s.MaxNotional.Sub(toUpBitDataStatic.Dec500)
}

func (s *Single) setExecuteParam(takeProfitPrice float64, twapSec float64) {
	s.twapSec = twapSec
	s.takeProfitPrice = takeProfitPrice
	s.closeDuration = time.Duration(twapSec) * time.Second
	toUpBitDataStatic.DyLog.GetLog().Infof("止盈价格: %.8f,平仓持续时间: %s,单账户上限:%s", takeProfitPrice, s.closeDuration.String(), s.MaxNotional)
}
