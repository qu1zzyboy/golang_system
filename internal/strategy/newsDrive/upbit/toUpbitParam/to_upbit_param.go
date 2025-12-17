package toUpbitParam

import "github.com/shopspring/decimal"

const MaxAccount = 11

var (
	Dec03    decimal.Decimal                     //首次下 0.3
	Dec103   = decimal.RequireFromString("1.03") //第一次下单价格1.03倍
	QtyTotal decimal.Decimal                     //单次下单总金额
)

func SetParam(qty, dec003 float64) {
	QtyTotal = decimal.NewFromFloat(qty)
	Dec03 = decimal.NewFromFloat(dec003)
}
