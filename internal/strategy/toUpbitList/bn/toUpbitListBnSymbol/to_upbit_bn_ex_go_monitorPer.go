package toUpbitListBnSymbol

import (
	"time"

	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
)

func (s *Single) monitorPer(accountIndex uint8) {
	var i int
	defer func() {
		toUpBitDataStatic.DyLog.GetLog().Infof("账户[%d],探测[%d]次,协程结束", accountIndex, i)
	}()
	price := s.FirstPriceBuy.Mul(dec2).Truncate(s.pScale)
OUTER:
	for i = 0; i <= 230; i++ {
		select {
		case <-s.ctxStop.Done():
			toUpBitDataStatic.DyLog.GetLog().Infof("收到关闭信号,退出探测协程")
			break OUTER
		default:
			//有成交或者本轮挂单成功
			if s.SecondArr[accountIndex].loadStop() || s.hasAllFilled.Load() {
				break OUTER
			}
			if err := bnOrderAppManager.GetMonitorManager().SendMonitorOrder(order_from, accountIndex, s.SymbolIndex,
				&orderModel.MyPlaceOrderReq{
					OrigPrice:     price,
					OrigVol:       s.PosTotalNeed,
					ClientOrderId: toUpBitDataStatic.GetClientOrderIdBy("sec-Mo"),
					StaticMeta:    s.StMeta,
					OrderType:     execute.ORDER_TYPE_LIMIT,
					OrderMode:     execute.ORDER_BUY_OPEN,
				}); err != nil {
				toUpBitDataStatic.DyLog.GetLog().Errorf("每秒探测订单失败: %v", err)
			}
			time.Sleep(300 * time.Microsecond) // 休眠 300 微秒
		}
	}
}
