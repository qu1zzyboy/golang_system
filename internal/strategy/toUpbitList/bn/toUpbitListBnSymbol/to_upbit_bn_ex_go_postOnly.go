package toUpbitListBnSymbol

import (
	"github.com/shopspring/decimal"
)

var (
	dec03    decimal.Decimal                     //首次下 0.3
	dec103   = decimal.RequireFromString("1.03") //第一次下单价格1.03倍
	qtyTotal decimal.Decimal                     //单次下单总金额
)

func SetParam(qty, dec003 float64) {
	qtyTotal = decimal.NewFromFloat(qty)
	dec03 = decimal.NewFromFloat(dec003)
}

/**
limit_maker协程,用到成员变量
posTotalNeed:在这里赋值,后面都只读
pScale: 多线程读安全
maxNotional: 初始化赋值之后都只读不写
secondArr: 不修改指针就安全
ctxStop:
hasAllFilled:
symbolIndex:
StMeta: 多线程读安全
firstPriceBuy:
thisOrderAccountId:

**/

/*
限制1: maxNotional 单账户单品种最大开仓上限
限制2：qtyTotal*0.3

退出循环:
1、账户没钱(不一定准)
*/
