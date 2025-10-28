package toUpbitListBnSymbol

import (
	"upbitBnServer/internal/cal/u64Cal"
	"upbitBnServer/internal/infra/observe/notify/notifyTg"
	"upbitBnServer/internal/quant/exchanges/binance/bnConst"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"

	"github.com/shopspring/decimal"
)

func (s *Single) checkPreOrder(markPrice_8 uint64) {
	if !s.hasInit {
		s.lastMarkPrice_8 = markPrice_8
		if err := s.initPreOrder(); err != nil {
			toUpBitListDataStatic.DyLog.GetLog().Errorf("%s 初始化交易对失败:  %s", s.StMeta.SymbolName, err.Error())
			return
		}
		//下小订单
		thisMarkPriceDec := decimal.New(int64(markPrice_8), -bnConst.PScale_8)
		if err := bnOrderAppManager.GetTradeManager().SendPlaceOrder(ws_req_from, s.preAccountKeyId, s.symbolIndex,
			&orderModel.MyPlaceOrderReq{
				OrigPrice:     thisMarkPriceDec.Mul(s.smallPercent).Truncate(s.pScale),
				OrigVol:       s.orderNum,
				ClientOrderId: s.clientOrderIdSmall,
				StaticMeta:    s.StMeta,
				OrderType:     execute.ORDER_TYPE_LIMIT,
				OrderMode:     execute.ORDER_SELL_OPEN,
			}); err != nil {
			toUpBitListDataStatic.DyLog.GetLog().Errorf("创建小订单错误: %s", err.Error())
		}
		return
	}
	// 价格距离上次变动有1%以上,开始移动订单
	if u64Cal.IsDiffMoreThanPercent100(markPrice_8, s.lastMarkPrice_8, 1) {
		// 缩小10的8次方倍
		thisMarkPriceDec := decimal.New(int64(markPrice_8), -bnConst.PScale_8)
		symbolKey := symbolStatic.GetSymbol().GetSymbolKey(s.StMeta.SymbolKeyId)
		modifySmallPrice := thisMarkPriceDec.Mul(s.smallPercent).Truncate(s.pScale)
		s.lastMarkPrice_8 = markPrice_8
		if _, ok := clientOrders.Load(s.clientOrderIdSmall); ok {
			toUpBitListDataStatic.DyLog.GetLog().Infof("[%d,%s] 触发[%d,%d,%d_%d],准备更新预挂单:%s",
				s.preAccountKeyId, symbolKey, s.lastMarkPrice_8, markPrice_8, s.pScale, s.qScale, modifySmallPrice)
			//更新订单价格
			if err := bnOrderAppManager.GetTradeManager().SendModifyOrder(ws_req_from, s.preAccountKeyId,
				&orderModel.MyModifyOrderReq{
					ModifyPrice:   modifySmallPrice,
					OrigVol:       s.orderNum,
					StaticMeta:    s.StMeta,
					ClientOrderId: s.clientOrderIdSmall,
					OrderMode:     execute.ORDER_SELL_OPEN,
				}); err != nil {
				notifyTg.GetTg().SendToUpBitMsg(map[string]string{
					"symbol": symbolKey,
					"op":     "更新5%订单失败",
					"error":  err.Error(),
				})
				toUpBitListDataStatic.DyLog.GetLog().Errorf("[%d] %s修改小订单错误: %s", s.preAccountKeyId, symbolKey, err.Error())
			}
		} else {
			toUpBitListDataStatic.DyLog.GetLog().Infof("[%d,%s] 触发[%d,%d,%d],准备重下预挂单:%s", s.preAccountKeyId, symbolKey, s.lastMarkPrice_8, markPrice_8, s.pScale, modifySmallPrice)
			if _, ok := clientOrderSig.Load(s.clientOrderIdSmall); ok {
				toUpBitListDataStatic.DyLog.GetLog().Infof("订单限流中,无法下单:%s", s.clientOrderIdSmall)
				return
			}
			//下小订单
			if err := bnOrderAppManager.GetTradeManager().SendPlaceOrder(ws_req_from, s.preAccountKeyId, s.symbolIndex,
				&orderModel.MyPlaceOrderReq{
					OrigPrice:     modifySmallPrice,
					OrigVol:       s.orderNum,
					ClientOrderId: s.clientOrderIdSmall,
					StaticMeta:    s.StMeta,
					OrderType:     execute.ORDER_TYPE_LIMIT,
					OrderMode:     execute.ORDER_SELL_OPEN,
				}); err != nil {
				notifyTg.GetTg().SendToUpBitMsg(map[string]string{
					"symbol": symbolKey,
					"op":     "预挂5%订单失败",
					"error":  err.Error(),
				})
				toUpBitListDataStatic.DyLog.GetLog().Errorf("%s创建小订单错误: %s", symbolKey, err.Error())
			}
		}
	}
}
