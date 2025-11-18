package bnOrderSdkModel

type orderSide uint8

const (
	sideBuy orderSide = iota
	sideSell
)

type positionSide uint8

const (
	positionSideLONG positionSide = iota
	positionSideSHORT
)

type orderType uint8

const (
	orderTypeLimit orderType = iota
	orderTypeMarket
	orderTypeLimitMaker
	orderTypeStopMarket
	orderTypeStop
	orderTypeTakeProfitMarket
	orderTypeTakeProfit
)

type timeInForce uint8

const (
	timeInForceGTC timeInForce = iota
	timeInForceGTX
	timeInForceIOC
	timeInForceFOK
	timeInForceGTD //最少要有效600s
)

var (
	orderSideArr    = [2]string{"BUY", "SELL"}
	positionSideArr = [2]string{"LONG", "SHORT"}
	orderTypeArr    = [7]string{"LIMIT", "MARKET", "LIMIT_MAKER", "STOP_MARKET", "STOP", "TAKE_PROFIT_MARKET", "TAKE_PROFIT"}
	timeInForceArr  = [5]string{"GTC", "GTX", "IOC", "FOK", "GTD"}
)
