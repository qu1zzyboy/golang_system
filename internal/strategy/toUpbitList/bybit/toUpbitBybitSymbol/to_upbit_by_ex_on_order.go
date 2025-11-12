package toUpbitBybitSymbol

import (
	"fmt"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListPos"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitParam"
)

func (s *Single) onFailureOrder(accountKeyId uint8, errCode [5]byte) {
	// {"code":-2019,"msg":"Margin is insufficient."},账户没钱,停止这一秒抽奖
	// switch {
	// case errCode[0] == '2' && errCode[1] == '0' && errCode[2] == '1' && errCode[3] == '9':
	// 	s.secondArr[accountKeyId].receiveStopBuy(accountKeyId)
	// }
}

func (s *Single) onSuccessOrder(evt toUpBitListDataAfter.OnSuccessEvt) {
	accountKeyId := evt.AccountKeyId
	if evt.IsOnline {
		s.clientOrderIds.Store(evt.ClientOrderId, accountKeyId)
	} else {
		if evt.Volume > 0 {
			if s.pos == nil {
				s.pos = toUpbitListPos.NewPosCal()
			}
			// 有成交更新可用仓位
			if evt.OrderMode.IsOpen() {
				posTotalAmount := s.pos.OpenFilled(accountKeyId, evt.Volume)
				// 判断是否完全开满
				if posTotalAmount >= s.posTotalNeed {
					s.hasAllFilled.Store(true)
					toUpBitDataStatic.DyLog.GetLog().Infof("完全成交,当前总仓位:%.8f,需要:%.8f", posTotalAmount, s.posTotalNeed)
				} else {
					left := s.posTotalNeed - posTotalAmount
					val := s.bidPrice.Load()
					if val != nil {
						if left*val.(float64) < toUpbitParam.Dec500 {
							s.hasAllFilled.Store(true)
							toUpBitDataStatic.DyLog.GetLog().Infof("完全成交,当前总仓位:%.8f,需要:%.8f", posTotalAmount, s.posTotalNeed)
						}
					}
				}
				toUpBitDataStatic.SendToUpBitMsg("发送开仓成交失败", map[string]string{
					"symbol": s.symbolName,
					"op":     fmt.Sprintf("账户[%d][%s]开仓成交:%.8f,当前总仓位:%.8f", accountKeyId, evt.ClientOrderId, evt.Volume, posTotalAmount),
				})
			} else {
				posTotalAmount := s.pos.CloseFilled(accountKeyId, evt.Volume)
				toUpBitDataStatic.SendToUpBitMsg("发送平仓成交失败", map[string]string{
					"symbol": s.symbolName,
					"op":     fmt.Sprintf("账户[%d][%s]平仓成交:%.8f,当前总仓位:%.8f", accountKeyId, evt.ClientOrderId, evt.Volume, posTotalAmount),
				})
			}
		}
		// 删除掉这个订单
		s.clientOrderIds.Delete(evt.ClientOrderId)
		toUpBitDataStatic.DyLog.GetLog().Infof("账户[%d][%s]订单完成[%.8f],剩余挂单数:%d", accountKeyId, string(evt.ClientOrderId[:]), evt.Volume, s.clientOrderIds.Length())
	}
}
