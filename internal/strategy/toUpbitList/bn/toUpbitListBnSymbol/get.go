package toUpbitListBnSymbol

import (
	"context"

	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"upbitBnServer/internal/strategy/toUpbitParam"
)

func GetParam(isMeme bool, symbolIndex int, cap float64) (gainPct, twapSec float64, err error) {
	resp, err := toUpbitParam.GetService().Compute(context.Background(), toUpbitParam.ComputeRequest{
		IsMeme:      isMeme,
		SymbolIndex: symbolIndex,
		MarketCapM:  cap,
	})
	if err != nil {
		return 0, 0, err
	}
	toUpBitListDataStatic.DyLog.GetLog().Debugf("param service result: symbol=%d gain=%.2f twap=%.2f", symbolIndex, resp.GainPct, resp.TwapSec)
	return resp.GainPct, resp.TwapSec, nil
}
