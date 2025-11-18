package toUpbitListBnSymbol

import (
	"time"
	"upbitBnServer/internal/quant/exchanges/binance/order/bnOrderAppManager"

	"upbitBnServer/internal/cal/u64Cal"
	"upbitBnServer/internal/infra/systemx/instanceEnum"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/pkg/utils/time2str"
)

func (s *Single) monitorPer(accountIndex uint8) {
	var i int
	defer func() {
		toUpBitDataStatic.DyLog.GetLog().Infof("账户[%d],探测[%d]次,协程结束", accountIndex, i)
	}()
	price := u64Cal.FromF64(2*s.priceMaxBuy, s.pScale.Uint8())
	num := s.pre.PointNum
OUTER:
	for i = 0; i <= 230; i++ {
		select {
		case <-s.ctxStop.Done():
			toUpBitDataStatic.DyLog.GetLog().Infof("收到关闭信号,退出探测协程")
			break OUTER
		default:
			//有成交或者本轮挂单成功
			if s.secondArr[accountIndex].loadStop() || s.hasAllFilled.Load() {
				break OUTER
			}
			if err := bnOrderAppManager.GetMonitorManager().SendMonitorOrder(accountIndex, orderModel.MyPlaceOrderReq{
				SymbolName:    s.symbolName,
				ClientOrderId: time2str.GetNowTimeStampMicroSlice16(),
				Pvalue:        price,
				Qvalue:        num,
				Pscale:        s.pScale,
				Qscale:        s.qScale,
				OrderMode:     execute.BUY_OPEN_LIMIT_MAKER,
				SymbolIndex:   s.symbolIndex,
				SymbolLen:     s.symbolLen,
				ReqFrom:       instanceEnum.TO_UPBIT_LIST_BN,
				UsageFrom:     to_upbit_main,
			}); err != nil {
				toUpBitDataStatic.DyLog.GetLog().Errorf("每秒探测订单失败: %v", err)
			}
			time.Sleep(300 * time.Microsecond) // 休眠 300 微秒
		}
	}
}
