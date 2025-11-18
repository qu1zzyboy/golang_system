package toUpbitListBnSymbol

import (
	"time"
	"upbitBnServer/internal/quant/exchanges/binance/order/bnOrderAppManager"

	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/instanceEnum"
	"upbitBnServer/internal/infra/systemx/usageEnum"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
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
	s.clientOrderIds.Range(func(clientOrderId systemx.WsId16B, accountKeyId uint8) bool {
		if int32(accountKeyId) == accountPreId {
			count++
			s.can.RefreshClientOrderId(clientOrderId)
			if err := bnOrderAppManager.GetTradeManager().SendCancelOrderBy(s.can, instanceEnum.TO_UPBIT_LIST_BN, usageEnum.TO_UPBIT_CANCEL_TRANSFER, accountKeyId); err != nil {
				toUpBitDataStatic.DyLog.GetLog().Errorf("撤销订单[%s]失败: %v", clientOrderId, err)
			}
		}
		return true
	})
	toUpBitDataStatic.DyLog.GetLog().Infof("%d 开始查询 %d的可划转金额,撤单数:%d", i, accountPreId, count)
	// 没有撤单,直接查询
	if count == 0 {
		bnOrderAppManager.GetTradeManager().SendQueryAccountBalance(usageEnum.TO_UPBIT_CANCEL_TRANSFER, uint8(accountPreId))
	}
}
