package toUpbitBnMode

import (
	"context"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"upbitBnServer/internal/strategy/toUpbitParam"

	"github.com/shopspring/decimal"
)

var (
	Mode   ModeBehavior
	dec200 = decimal.NewFromFloat(200)
)

type ModeBehavior interface {
	IsPlacePreOrder() bool
	ShouldExitOnTakeProfit(priceBuy, takeProfit float64) bool
	IsDynamicStopLossTrig(bid, maxPriceF64 float64) bool
	GetTreeNewsFlag() bool
	GetTakeProfitParam(isMeme bool, symbolIndex int, cap float64) (gainPct, twapSec float64, err error)
	GetTransferAmount(amount decimal.Decimal) decimal.Decimal
}

type DebugMode struct{}

func (d DebugMode) GetTransferAmount(amount decimal.Decimal) decimal.Decimal {
	return amount
}

func (d DebugMode) GetTakeProfitParam(_ bool, _ int, _ float64) (float64, float64, error) {
	return 7, 10, nil
}

func (d DebugMode) GetTreeNewsFlag() bool {
	return true
}

func (d DebugMode) IsDynamicStopLossTrig(bid, maxPriceF64 float64) bool {
	return bid < maxPriceF64*0.86
}

func (d DebugMode) IsPlacePreOrder() bool { return false }

func (d DebugMode) ShouldExitOnTakeProfit(_, _ float64) bool { return false }

type MocaMode struct{}

func (l MocaMode) GetTransferAmount(amount decimal.Decimal) decimal.Decimal {
	// 留200u作为挂单保证金
	return amount.Sub(dec200)
}

func (l MocaMode) GetTakeProfitParam(isMeme bool, symbolIndex int, cap float64) (float64, float64, error) {
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

func (l MocaMode) GetTreeNewsFlag() bool {
	return false
}

func (l MocaMode) IsDynamicStopLossTrig(bid, maxPriceF64 float64) bool {
	return bid < maxPriceF64*0.95
}

func (l MocaMode) IsPlacePreOrder() bool { return true }

func (l MocaMode) ShouldExitOnTakeProfit(priceBuy, takeProfit float64) bool {
	return takeProfit > 0 && priceBuy > takeProfit
}

type LiveMode struct{}

func (l LiveMode) GetTransferAmount(amount decimal.Decimal) decimal.Decimal {
	// 留200u作为挂单保证金
	return amount.Sub(dec200)
}

func (l LiveMode) GetTakeProfitParam(isMeme bool, symbolIndex int, cap float64) (float64, float64, error) {
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

func (l LiveMode) GetTreeNewsFlag() bool {
	return false
}

func (l LiveMode) IsDynamicStopLossTrig(bid, maxPriceF64 float64) bool {
	return bid < maxPriceF64*0.95
}

func (l LiveMode) IsPlacePreOrder() bool { return true }

func (l LiveMode) ShouldExitOnTakeProfit(priceBuy, takeProfit float64) bool {
	return takeProfit > 0 && priceBuy > takeProfit
}
