package toUpBitListBnExecute

import (
	"time"

	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderModel"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/bn/toUpBitListBnAccount"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"github.com/shopspring/decimal"
)

const tranSpecial = orderBelongEnum.TO_UPBIT_LIST_LOOP_CANCEL_TRANSFER

var dec4 = decimal.NewFromInt(4)

func (s *Execute) cancelAndTransfer(i, accountPreId int32) {
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

func (s *Execute) OnCanceledOrder(accountKeyId uint8) {
	//接收到撤单的返回,开始查询转出账户的可划转余额
	_ = bnOrderAppManager.GetTradeManager().SendQueryAccountBalance(tranSpecial, accountKeyId)
}

func (s *Execute) OnMaxWithdrawAmount(accountKeyId uint8, maxWithdrawAmount decimal.Decimal) {
	//接收到可划转金额的返回,划入母账户
	_ = toUpBitListBnAccount.GetBnAccountManager().TransferIn(int32(accountKeyId), maxWithdrawAmount)
}

func (s *Execute) OnTransOut(maxWithdrawAmount decimal.Decimal) {
	//接受到母账户金额的返回,从母账户划出
	var err error
	var accountKeyId int32 = 0
	for {
		accountKeyId = s.toAccountId.Load()
		if err = toUpBitListBnAccount.GetBnAccountManager().TransferOut(accountKeyId, maxWithdrawAmount); err == nil {
			s.maxNotionalArr[accountKeyId].Store(decimal.Min(maxWithdrawAmount.Mul(dec4), s.maxNotional))
			break
		} else {
			toUpBitListDataStatic.DyLog.GetLog().Errorf("划转到[%d]失败:%v", accountKeyId, err)
		}
	}
}
