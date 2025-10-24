package toUpbitListBnSymbol

import (
	"context"

	"upbitBnServer/internal/strategy/params"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
)

func GetParam(isMeme bool, cap float64, symbolName string) (gainPct, twapSec float64, err error) {
	resp, err := params.GetService().Compute(context.Background(), params.ComputeRequest{
		MarketCapM: cap,
		IsMeme:     isMeme,
		SymbolName: symbolName,
	})
	if err != nil {
		return 0, 0, err
	}
	toUpBitListDataStatic.DyLog.GetLog().Debugf("param service result: symbol=%s gain=%.2f twap=%.2f", symbolName, resp.GainPct, resp.TwapSec)
	return resp.GainPct, resp.TwapSec, nil
}
