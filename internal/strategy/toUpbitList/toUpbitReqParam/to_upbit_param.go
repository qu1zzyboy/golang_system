package toUpbitReqParam

import "github.com/shopspring/decimal"

const MaxAccount = 11

var (
	Dec03         decimal.Decimal //每次抽奖的资金比例
	QtyTotal      decimal.Decimal //单次下单总金额
	Dec500        decimal.Decimal //完全成交判定阈值 500u
	PriceRiceTrig float64         // 价格触发阈值,当价格变化超过该值时触发
)

func SetParam(qty, dec003, priceRiceTrig float64, dec500 int64) {
	QtyTotal = decimal.NewFromFloat(qty)
	Dec03 = decimal.NewFromFloat(dec003)
	Dec500 = decimal.NewFromInt(dec500)
	PriceRiceTrig = priceRiceTrig
}
