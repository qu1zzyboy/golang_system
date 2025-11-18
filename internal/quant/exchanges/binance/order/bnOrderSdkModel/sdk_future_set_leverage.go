package bnOrderSdkModel

import (
	"fmt"

	"upbitBnServer/pkg/utils/convertx"
	"upbitBnServer/pkg/utils/timeUtils"
)

// FutureLeverage
type FutureLeverageSdk struct {
	Symbol   string `json:"symbol"`   //YES	交易对
	Leverage uint8  `json:"leverage"` //YES	目标杠杆倍数：1 到 125 整数
}

func (api *FutureLeverageSdk) Symbol_(symbol string) *FutureLeverageSdk {
	api.Symbol = symbol
	return api
}
func (api *FutureLeverageSdk) Leverage_(leverage uint8) *FutureLeverageSdk {
	api.Leverage = leverage
	return api
}

var b_leverage = []byte("&leverage=")

func (api *FutureLeverageSdk) ParseRestReq() []byte {
	var orig []byte
	orig = append(orig, b_SYMBOL...)
	orig = append(orig, api.Symbol...)

	orig = append(orig, b_leverage...)
	orig = append(orig, fmt.Sprintf("%d", api.Leverage)...)

	orig = append(orig, b_TIME_STAMP...)
	orig = convertx.AppendValueToBytes(orig, timeUtils.GetNowTimeUnixMilli())
	return orig
}
