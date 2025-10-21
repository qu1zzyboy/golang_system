package orderSdkBnModel

import (
	"github.com/hhh500/quantGoInfra/pkg/utils/convertx"
	"github.com/hhh500/quantGoInfra/pkg/utils/timeUtils"
	"github.com/hhh500/upbitBnServer/internal/quant/account/universalTransfer"
	"github.com/shopspring/decimal"
)

type UniversalTransferSdk struct {
	Amount          decimal.Decimal `json:"amount"`          //YES
	FromAccountType string          `json:"fromAccountType"` //YES	"SPOT","USDT_FUTURE","COIN_FUTURE","MARGIN"(Cross),"ISOLATED_MARGIN"
	ToAccountType   string          `json:"toAccountType"`   //YES	"SPOT","USDT_FUTURE","COIN_FUTURE","MARGIN"(Cross),"ISOLATED_MARGIN"
	Asset           string          `json:"asset"`           //YES
	ToEmail         string          `json:"toEmail"`         //NO
	FromEmail       string          `json:"fromEmail"`       //NO
}

func (api *UniversalTransferSdk) FromEmail_(fromEmail string) *UniversalTransferSdk {
	api.FromEmail = fromEmail
	return api
}
func (api *UniversalTransferSdk) ToEmail_(toEmail string) *UniversalTransferSdk {
	api.ToEmail = toEmail
	return api
}
func (api *UniversalTransferSdk) FromAccountType_(fromAccountType string) *UniversalTransferSdk {
	api.FromAccountType = fromAccountType
	return api
}
func (api *UniversalTransferSdk) ToAccountType_(toAccountType string) *UniversalTransferSdk {
	api.ToAccountType = toAccountType
	return api
}

func (api *UniversalTransferSdk) Asset_(asset string) *UniversalTransferSdk {
	api.Asset = asset
	return api
}
func (api *UniversalTransferSdk) Amount_(amount decimal.Decimal) *UniversalTransferSdk {
	api.Amount = amount
	return api
}

var (
	b_FROM_EMAIL       = []byte("fromEmail=")
	b_TO_EMAIL         = []byte("&toEmail=")
	b_AMOUNT           = []byte("&amount=")
	b_FROM_ACCOUNTTYPE = []byte("&fromAccountType=")
	b_TO_ACCOUNTTYPE   = []byte("&toAccountType=")
	b_ASSET            = []byte("&asset=")
)

func (api *UniversalTransferSdk) ParseRestReq() []byte {
	var orig []byte
	orig = append(orig, b_FROM_EMAIL...)
	orig = append(orig, api.FromEmail...)

	orig = append(orig, b_TO_EMAIL...)
	orig = append(orig, api.ToEmail...)

	orig = append(orig, b_FROM_ACCOUNTTYPE...)
	orig = append(orig, api.FromAccountType...)

	orig = append(orig, b_TO_ACCOUNTTYPE...)
	orig = append(orig, api.ToAccountType...)

	orig = append(orig, b_ASSET...)
	orig = append(orig, api.Asset...)

	orig = append(orig, b_AMOUNT...)
	orig = append(orig, api.Amount.String()...)

	orig = append(orig, b_TIME_STAMP...)
	orig = convertx.AppendValueToBytes(orig, timeUtils.GetNowTimeUnixMilli())
	return orig
}

func newUniversalTransferSdk() *UniversalTransferSdk {
	return &UniversalTransferSdk{}
}

func GetSpotTransferSdk(req *universalTransfer.UniversalTransferReq) *UniversalTransferSdk {
	return newUniversalTransferSdk().FromEmail_(req.From).
		ToEmail_(req.To).
		FromAccountType_(req.FromAcType).
		ToAccountType_(req.ToAcType).
		Asset_(req.Asset).
		Amount_(req.Amount)
}
