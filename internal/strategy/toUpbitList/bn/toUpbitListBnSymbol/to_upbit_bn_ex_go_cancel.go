package toUpbitListBnSymbol

import (
	"time"

	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderModel"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
)

/**
撤单协程,用到成员变量
clientOrderIds:多线程安全
StMeta: 多线程读安全
**/

func (s *Single) cancelAndTransfer(i, accountPreId int32) {
	// 从序号2开始撤销订单,需要保证订单状态的维护
	if i < 2 {
		return
	}
	if i == 2 {
		accountPreId = 0
	}
	time.Sleep(400 * time.Millisecond)

	// 已经撤单的数量
	count := 0
	s.clientOrderIds.Range(func(clientOrderId string, accountKeyId uint8) bool {
		if int32(accountKeyId) == accountPreId {
			count++
			if err := bnOrderAppManager.GetTradeManager().SendCancelOrder(tranSpecial, accountKeyId,
				&orderModel.MyQueryOrderReq{
					ClientOrderId: clientOrderId,
					StaticMeta:    s.StMeta,
				}); err != nil {
				toUpBitListDataStatic.DyLog.GetLog().Errorf("撤销订单[%s]失败: %v", clientOrderId, err)
			}
		}
		return true
	})
	// 没有撤单,直接查询
	if count == 0 {
		bnOrderAppManager.GetTradeManager().SendQueryAccountBalance(tranSpecial, uint8(accountPreId))
	}
}
