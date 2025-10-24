package orderSdkBnModel

import (
	"upbitBnServer/pkg/utils/convertx"
	"upbitBnServer/pkg/utils/timeUtils"
)

// FutureLeverage
type FutureSymbolConfigSdk struct {
}

func (api *FutureSymbolConfigSdk) ParseRestReq() []byte {
	var orig []byte
	orig = append(orig, b_TIME_STAMP...)
	orig = convertx.AppendValueToBytes(orig, timeUtils.GetNowTimeUnixMilli())
	return orig
}

func NewFutureSymbolConfigSdk() *FutureSymbolConfigSdk {
	return &FutureSymbolConfigSdk{}
}
