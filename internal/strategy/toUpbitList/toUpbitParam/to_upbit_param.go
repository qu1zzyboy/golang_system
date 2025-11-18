package toUpbitParam

const MaxAccount = 6

var (
	F03        float64 //每次抽奖的资金比例
	QtyTotal   float64 //单次下单总金额
	Dec500     float64 //完全成交判定阈值 500u
	AccountLen int     //当前系统内部的账户总数
)

func SetParam(qty, dec003, dec500 float64, accountLen int) {
	QtyTotal = qty
	F03 = dec003
	Dec500 = dec500
	AccountLen = accountLen
}
