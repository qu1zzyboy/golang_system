package orderSdkBnModel

import (
	"upbitBnServer/pkg/utils/convertx"
	"upbitBnServer/pkg/utils/timeUtils"
)

// FutureLeverage
type FutureTradeingRuleSdk struct {
	SymbolName string // 交易对名称
}

func (api *FutureTradeingRuleSdk) ParseRestReq() []byte {
	var orig []byte
	if api.SymbolName != "" {
		orig = append(orig, b_SYMBOL...)
		orig = append(orig, api.SymbolName...)
	}
	orig = append(orig, b_TIME_STAMP...)
	orig = convertx.AppendValueToBytes(orig, timeUtils.GetNowTimeUnixMilli())
	return orig
}

func NewFutureTradeingRuleSdk() *FutureTradeingRuleSdk {
	return &FutureTradeingRuleSdk{}
}
