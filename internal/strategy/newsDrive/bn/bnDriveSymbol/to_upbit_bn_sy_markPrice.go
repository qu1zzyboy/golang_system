package bnDriveSymbol

import (
	"upbitBnServer/internal/strategy/newsDrive/common/driverStatic"

	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/pkg/container/map/myMap"
	"upbitBnServer/pkg/container/pool/byteBufPool"

	"github.com/shopspring/decimal"
)

var (
	dec12              = decimal.RequireFromString("12.00")     //2倍最小下单金额
	dec2               = decimal.RequireFromString("2.00")      //2倍最小下单金额
	dec5               = decimal.RequireFromString("0.33")      //小订单比例
	dec1               = decimal.RequireFromString("1.0")       //1.0
	clientOrders       = myMap.NewMySyncMap[string, struct{}]() //clientOrderId-->占位符,所有的挂单状态的订单
	clientOrderSig     = myMap.NewMySyncMap[string, struct{}]() //clientOrderId-->占位符,有就不下单
	ClientOrderIsCheck = myMap.NewMySyncMap[string, struct{}]() //clientOrderId-->占位符,有就不检查
)

func (s *Single) onMarkPrice(len int, bufPtr *[]byte) {
	defer byteBufPool.ReleaseBuffer(bufPtr)
	data := (*bufPtr)[:len]
	if toUpBitListDataAfter.LoadTrig() {
		maxBuy, ok := s.symbol.OnMarkPriceAfter(data)
		if ok {
			driverStatic.DyLog.GetLog().Infof("最新[max_buy:%.8f]标记价格: %s", maxBuy, string(data))
		}
	} else {
		thisMarkPrice := s.symbol.OnMarkPriceBefore(data)
		s.pre.CheckPreOrder(thisMarkPrice)
	}
}
