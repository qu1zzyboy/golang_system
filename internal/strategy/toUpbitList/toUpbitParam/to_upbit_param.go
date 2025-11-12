package toUpbitParam

const MaxAccount = 6

var (
	F03      float64 //每次抽奖的资金比例
	QtyTotal float64 //单次下单总金额
	Dec500   float64 //完全成交判定阈值 500u
)

func SetParam(qty, dec003, dec500 float64) {
	QtyTotal = qty
	F03 = dec003
	Dec500 = dec500
}
