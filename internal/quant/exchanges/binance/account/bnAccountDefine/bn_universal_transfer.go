package bnAccountDefine

import (
	"github.com/shopspring/decimal"
)

type UniversalTransferReq struct {
	From       string          //转出的账户,默认从母账户转出
	To         string          //转入的账户
	FromAcType string          //转出的账户类型,SPOT,
	ToAcType   string          //转入的账户类型
	Asset      string          //划转的资产名称
	Amount     decimal.Decimal //划转的数量
}

type FromType string

const (
	SPOT            FromType = "SPOT"            //币币账户
	USDT_FUTURE     FromType = "USDT_FUTURE"     //USDT合约账户
	COIN_FUTURE     FromType = "COIN_FUTURE"     //币本位合约账户
	MARGIN          FromType = "MARGIN"          //杠杆账户
	ISOLATED_MARGIN FromType = "ISOLATED_MARGIN" //逐仓杠杆账户
)
