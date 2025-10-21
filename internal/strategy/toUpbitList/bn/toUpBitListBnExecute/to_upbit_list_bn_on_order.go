package toUpBitListBnExecute

import (
	"fmt"

	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"github.com/shopspring/decimal"
)

func (s *Execute) OnFailureOrder(accountKeyId uint8, errCode int64) {
	// {"code":-2019,"msg":"Margin is insufficient."},账户没钱,停止这一秒抽奖
	if errCode == -2019 {
		if s.hasInToSecondPerLoopArr[accountKeyId].Load() {
			if !s.stopThisSecondPerArr[accountKeyId].Load() {
				s.stopThisSecondPerArr[accountKeyId].Store(true)
				toUpBitListDataStatic.DyLog.GetLog().Infof("账户[%d]没钱,停止这一秒抽奖", accountKeyId)
			}
		}
	}
}

func (s *Execute) OnSuccessOrder(evt toUpBitListDataAfter.OnSuccessEvt) {
	accountKeyId := evt.AccountKeyId
	if evt.IsOnline {
		//挂单可能有两个来源,只消费一次
		if _, ok := s.clientOrderIds.Load(evt.ClientOrderId); ok {
			return
		}
		s.clientOrderIds.Store(evt.ClientOrderId, accountKeyId)
		// if evt.OrderMode.IsOpen() {
		// 	// 买入开多挂单成功,停止这一秒抽奖
		// 	if s.hasInToSecondPerLoopArr[accountKeyId].Load() {
		// 		s.stopThisSecondPerArr[accountKeyId].Store(true)
		// 		toUpBitListDataStatic.DyLog.GetLog().Infof("[%d][%s][%d]挂单成功,停止这一秒抽奖", accountKeyId, evt.ClientOrderId, evt.TimeStamp)
		// 	}
		// }
	} else {
		// 非挂单只有一个来源
		if evt.Volume.GreaterThan(decimal.Zero) {
			// 有成交更新可用仓位
			if evt.OrderMode.IsOpen() {
				posTotalAmount := s.pos.OpenFilled(accountKeyId, evt.Volume)
				// 判断是否完全开满
				lastBid64, ok := toUpBitListDataAfter.LoadBidPrice()
				if ok {
					left := s.posTotalNeed.Sub(posTotalAmount).Abs()
					lastPrice := decimal.NewFromFloat(lastBid64)
					if left.Mul(lastPrice).LessThan(toUpBitListDataStatic.Dec500) {
						s.hasAllFilled.Store(true)
						toUpBitListDataStatic.DyLog.GetLog().Infof("完全成交,当前总仓位:%s,需要:%s", posTotalAmount, s.posTotalNeed)
					}
				}
				toUpBitListDataStatic.SendToUpBitMsg("发送开仓成交失败", map[string]string{
					"symbol": toUpBitListDataAfter.TrigSymbolName,
					"op":     fmt.Sprintf("账户[%d][%s]开仓成交:%s,当前总仓位:%s", accountKeyId, evt.ClientOrderId, evt.Volume, posTotalAmount),
				})
			} else {
				posTotalAmount := s.pos.CloseFilled(accountKeyId, evt.Volume)
				toUpBitListDataStatic.SendToUpBitMsg("发送平仓成交失败", map[string]string{
					"symbol": toUpBitListDataAfter.TrigSymbolName,
					"op":     fmt.Sprintf("账户[%d][%s]平仓成交:%s,当前总仓位:%s", accountKeyId, evt.ClientOrderId, evt.Volume, posTotalAmount),
				})
			}
		}
		// 删除掉这个订单
		s.clientOrderIds.Delete(evt.ClientOrderId)
		toUpBitListDataStatic.DyLog.GetLog().Infof("账户[%d][%s]订单完成[%s],剩余挂单数:%d",
			accountKeyId, evt.ClientOrderId, evt.Volume, s.clientOrderIds.Length())
	}
}
