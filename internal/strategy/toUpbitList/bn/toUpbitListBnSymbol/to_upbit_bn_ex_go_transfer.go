package toUpbitListBnSymbol

import (
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpBitListBnAccount"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"

	"github.com/shopspring/decimal"
)

const tranSpecial = orderBelongEnum.TO_UPBIT_LIST_LOOP_CANCEL_TRANSFER

var dec4 = decimal.NewFromInt(4)

func (s *Single) onCanceledOrder(accountKeyId uint8) {
	//接收到撤单的返回,开始查询转出账户的可划转余额
	_ = bnOrderAppManager.GetTradeManager().SendQueryAccountBalance(tranSpecial, accountKeyId)
}

func (s *Single) onMaxWithdrawAmount(accountKeyId uint8, maxWithdrawAmount decimal.Decimal) {
	//接收到可划转金额的返回,划入母账户
	_ = toUpBitListBnAccount.GetBnAccountManager().TransferIn(int32(accountKeyId), maxWithdrawAmount)
}

func (s *Single) OnTransOut(maxWithdrawAmount decimal.Decimal) {
	//这里必须要多线程,没有办法在这个主线程内等待划转返回

	safex.SafeGo("on_trans_out", func() {
		//接受到母账户金额的返回,从母账户划出
		var err error
		var accountKeyId int32 = 0
		for {
			accountKeyId = s.toAccountId.Load()
			if err = toUpBitListBnAccount.GetBnAccountManager().TransferOut(accountKeyId, maxWithdrawAmount); err == nil {
				s.secondArr[accountKeyId].maxNotional.Store(decimal.Min(maxWithdrawAmount.Mul(dec4), s.maxNotional))
				break
			} else {
				toUpBitDataStatic.DyLog.GetLog().Errorf("划转到[%d]失败:%v", accountKeyId, err)
			}
		}
	})
}
