package toUpbitListBnSymbol

import (
	"context"

	"upbitBnServer/internal/strategy/params"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
)

func GetParam(isMeme bool, symbolIndex int, cap float64) (gainPct, twapSec float64, err error) {
	resp, err := params.GetService().Compute(context.Background(), isMeme, 0, cap)
	if err != nil {
		return 0, 0, err
	}
	toUpBitListDataStatic.DyLog.GetLog().Debugf("param service result: symbol=%d gain=%.2f twap=%.2f", symbolIndex, resp.GainPct, resp.TwapSec)
	return resp.GainPct, resp.TwapSec, nil
}
