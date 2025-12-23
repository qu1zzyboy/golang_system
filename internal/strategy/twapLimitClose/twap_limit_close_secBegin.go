package twapLimitClose

import (
	"upbitBnServer/internal/infra/observe/notify/notifyTg"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/quant/execute/order/orderStatic"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/pkg/container/map/myMap"

	"github.com/shopspring/decimal"
)

func InitPerSecondBegin(orderMode execute.MyOrderMode, accountKeyId uint8, symbolIndex int, pScale, qScale int32, stMeta *symbolStatic.StaticTrade, closeOrderIds *myMap.MySyncMap[string, bool],
	priceDec, accountPos, closePer decimal.Decimal) {

	if accountPos.IsZero() {
		toUpBitDataStatic.DyLog.GetLog().Infof("[%s],当前仓位为0,[%s]", stMeta.SymbolName, accountPos)
		return
	}
	var (
		closeN = 100
	)
	if closePer.IsZero() {
		// 防止 0 数量订单
		closeN = 1
		closePer = accountPos
	}

	dec1 := decimal.NewFromFloat(1.00)
	risePercent := decimal.NewFromFloat(0.001)

	for i := 0; i < closeN; i++ {
		var price decimal.Decimal
		if orderMode.IsBuy() {
			price = priceDec.Mul(dec1.Sub(risePercent.Mul(decimal.NewFromInt(int64(i + 1))))) // 每次加0.1%
		} else {
			price = priceDec.Mul(dec1.Add(risePercent.Mul(decimal.NewFromInt(int64(i + 1))))) // 每次加0.1%
		}
		num := closePer
		if accountPos.LessThan(closePer) {
			num = accountPos // 吃掉剩余
		}
		if accountPos.LessThanOrEqual(decimal.Zero) {
			break
		}
		clientOrderId := toUpBitDataStatic.GetClientOrderIdBy("server_twap")
		if err := bnOrderAppManager.GetTradeManager().SendPlaceOrder(orderBelongEnum.TO_UPBIT_LIST_PRE, accountKeyId, symbolIndex, &orderModel.MyPlaceOrderReq{
			OrigPrice:     price.Truncate(pScale),
			OrigVol:       num.Truncate(qScale),
			ClientOrderId: clientOrderId,
			StaticMeta:    stMeta,
			OrderType:     execute.ORDER_TYPE_LIMIT,
			OrderMode:     orderMode,
		}); err != nil {
			toUpBitDataStatic.DyLog.GetLog().Errorf("每秒平仓创建订单失败: %v", err)
		}
		closeOrderIds.Store(clientOrderId, false)
		accountPos = accountPos.Sub(num)
	}
}

func RefreshPerSecondBegin(orderMode execute.MyOrderMode, accountKeyId uint8, pScale int32, stMeta *symbolStatic.StaticTrade, closeOrderIds *myMap.MySyncMap[string, bool],
	priceDec decimal.Decimal) {

	dec1 := decimal.NewFromFloat(1.00)
	risePercent := decimal.NewFromFloat(0.001)

	var i = 0
	closeOrderIds.Range(func(clientOrderId string, v bool) bool {
		if v {
			return true
		}
		oMeta, ok := orderStatic.GetService().GetOrderMeta(clientOrderId)
		// 不属于这个服务的订单直接 pass
		if !ok {
			toUpBitDataStatic.DyLog.GetLog().Errorf("[%d] 每秒开始REFRESH_ORDER: [%s,%s] not found", accountKeyId, stMeta.SymbolName, clientOrderId)
			return true
		}
		i++
		var price decimal.Decimal
		if orderMode.IsBuy() {
			price = priceDec.Mul(dec1.Sub(risePercent.Mul(decimal.NewFromInt(int64(i))))) // 每次加0.1%
		} else {
			price = priceDec.Mul(dec1.Add(risePercent.Mul(decimal.NewFromInt(int64(i))))) // 每次加0.1%
		}
		if err := bnOrderAppManager.GetTradeManager().SendModifyOrder(orderBelongEnum.TO_UPBIT_LIST_PRE, accountKeyId, &orderModel.MyModifyOrderReq{
			ModifyPrice:   price.Truncate(pScale),
			OrigVol:       oMeta.OrigVolume,
			StaticMeta:    stMeta,
			ClientOrderId: clientOrderId,
			OrderMode:     orderMode,
		}); err != nil {
			notifyTg.GetTg().SendToUpBitMsg(map[string]string{
				"symbol": stMeta.SymbolName,
				"op":     "更新twapLimitClose订单失败",
				"error":  err.Error(),
			})
			toUpBitDataStatic.DyLog.GetLog().Errorf("%s修改twapLimitClose订单错误: %s", stMeta.SymbolName, err.Error())
		}
		closeOrderIds.Store(clientOrderId, false) // 重置为 false
		return true
	})
}
