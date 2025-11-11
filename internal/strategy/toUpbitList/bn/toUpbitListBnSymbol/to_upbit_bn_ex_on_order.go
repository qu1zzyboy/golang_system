package toUpbitListBnSymbol

import (
	"fmt"

	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"

	"github.com/shopspring/decimal"
)

func (s *Single) onFailureOrder(accountKeyId uint8, errCode int64) {
	// {"code":-2019,"msg":"Margin is insufficient."},账户没钱,停止这一秒抽奖
	if errCode == -2019 {
		// s.secondArr[accountKeyId].receiveStop(accountKeyId)
	}
}

func (s *Single) onSuccessOrder(evt toUpBitListDataAfter.OnSuccessEvt) {
	accountKeyId := evt.AccountKeyId
	if evt.IsOnline {
		s.clientOrderIds.Store(evt.ClientOrderId, accountKeyId)
	} else {
		if evt.Volume.GreaterThan(decimal.Zero) {
			// 有成交更新可用仓位
			if evt.OrderMode.IsOpen() {
				posTotalAmount := s.pos.OpenFilled(accountKeyId, evt.Volume)
				// 判断是否完全开满
				left := s.posTotalNeed.Sub(posTotalAmount).Abs()

				val := s.bidPrice.Load()
				if val != nil {
					lastPrice := decimal.NewFromFloat(val.(float64))
					if left.Mul(lastPrice).LessThan(toUpBitDataStatic.Dec500) {
						s.hasAllFilled.Store(true)
						toUpBitDataStatic.DyLog.GetLog().Infof("完全成交,当前总仓位:%s,需要:%s", posTotalAmount, s.posTotalNeed)
					}
				}
				toUpBitDataStatic.SendToUpBitMsg("发送开仓成交失败", map[string]string{
					"symbol": s.StMeta.SymbolName,
					"op":     fmt.Sprintf("账户[%d][%s]开仓成交:%s,当前总仓位:%s", accountKeyId, evt.ClientOrderId, evt.Volume, posTotalAmount),
				})
			} else {
				posTotalAmount := s.pos.CloseFilled(accountKeyId, evt.Volume)
				toUpBitDataStatic.SendToUpBitMsg("发送平仓成交失败", map[string]string{
					"symbol": s.StMeta.SymbolName,
					"op":     fmt.Sprintf("账户[%d][%s]平仓成交:%s,当前总仓位:%s", accountKeyId, evt.ClientOrderId, evt.Volume, posTotalAmount),
				})
			}
		}
		// 删除掉这个订单
		s.clientOrderIds.Delete(evt.ClientOrderId)
		toUpBitDataStatic.DyLog.GetLog().Infof("账户[%d][%s]订单完成[%s],剩余挂单数:%d",
			accountKeyId, evt.ClientOrderId, evt.Volume, s.clientOrderIds.Length())
	}
}
