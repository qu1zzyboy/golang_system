package twapLimitClose

import (
	"fmt"
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

func RefreshPerSecondEnd(orderMode execute.MyOrderMode, accountKeyId uint8, stMeta *symbolStatic.StaticTrade, closeOrderIds *myMap.MySyncMap[string, bool], priceDec decimal.Decimal) {
	if closeOrderIds.Length() <= 0 {
		toUpBitDataStatic.DyLog.GetLog().Errorf("账户[%d]平多单已全部成交，停止刷新止盈单", accountKeyId)
		return
	}
	clientOrderId, found := pickOneFalse(closeOrderIds)
	if !found {
		toUpBitDataStatic.DyLog.GetLog().Errorf("账户[%d]平多单 刷新止盈单失败，没有可刷新的订单", accountKeyId)
		return
	}

	oMeta, ok := orderStatic.GetService().GetOrderMeta(clientOrderId)
	// 不属于这个服务的订单直接 pass
	if !ok {
		toUpBitDataStatic.DyLog.GetLog().Errorf("[%d] 每秒REFRESH_ORDER: [%s,%s] not found", accountKeyId, stMeta.SymbolName, clientOrderId)
		return
	}
	fmt.Println("modify", clientOrderId)
	// 修改订单价格和数量,改单有最小下单数量限制
	if err := bnOrderAppManager.GetTradeManager().SendModifyOrder(orderBelongEnum.TO_UPBIT_LIST_PRE, accountKeyId, &orderModel.MyModifyOrderReq{
		ModifyPrice:   priceDec,
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
}

func pickOneFalse(closeOrderIds *myMap.MySyncMap[string, bool]) (string, bool) {
	var (
		result string
		found  bool
	)
	// ---------- 第一次遍历：找 false ----------
	closeOrderIds.Range(func(k string, v bool) bool {
		if !v {
			result = k
			found = true
			closeOrderIds.Store(k, true) // ⭐ 找到后立刻标记为 true
			return false
		}
		return true
	})

	if found {
		return result, true
	}
	// ---------- 没找到：全部重置为 false ----------
	closeOrderIds.Range(func(k string, v bool) bool {
		closeOrderIds.Store(k, false)
		return true
	})

	if closeOrderIds.Length() == 0 {
		return result, false
	}

	// ---------- 第二次遍历：重新找 false ----------
	closeOrderIds.Range(func(k string, v bool) bool {
		if !v {
			result = k
			found = true
			closeOrderIds.Store(k, true) // ⭐ reset 后同样要标记
			return false
		}
		return true
	})

	return result, found
}
