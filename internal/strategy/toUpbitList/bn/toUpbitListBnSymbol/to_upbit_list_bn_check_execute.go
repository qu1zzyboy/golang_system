package toUpbitListBnSymbol

import (
	"fmt"
	"time"

	"github.com/hhh500/quantGoInfra/quant/exchanges/binance/bnConst"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/bn/toUpBitListBnExecute"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"github.com/shopspring/decimal"
)

func (s *Single) IntoExecuteCheck(eventTs int64, trigFlag string, riseValue float64, priceTrig_8 uint64) {
	priceLimit_U10 := s.priceMax_10
	if !toUpBitListDataStatic.IsDebug {
		// 价格必须大于2min之前的价格
		if priceTrig_8 < s.last2MinClose_8 {
			toUpBitListDataStatic.SendToUpBitMsg("发送bn触发消息失败", map[string]string{
				"msg":  trigFlag + fmt.Sprintf("触发但价格小于2min之前[%d,%d]", s.last2MinClose_8, priceTrig_8),
				"bn品种": s.StMeta.SymbolName,
				"上涨幅度": fmt.Sprintf("%.2f%%", riseValue*100),
			})
			return
		}
		// 价格上限
		if priceLimit_U10 == 0 {
			toUpBitListDataStatic.SendToUpBitMsg("发送bn触发消息失败", map[string]string{
				"msg":  trigFlag + "触发但还未收到标记价格",
				"bn品种": s.StMeta.SymbolName,
				"上涨幅度": fmt.Sprintf("%.2f%%", riseValue*100),
			})
			return
		}
	}
	// 设置参数
	symbolName := s.StMeta.SymbolName
	toUpBitListDataAfter.Trig(symbolName, s.symbolIndex)

	if toUpBitListDataStatic.IsDebug {
		toUpBitListDataAfter.UpdateTreeNewsFlag()           //实盘待删除
		toUpBitListBnExecute.GetExecute().ReceiveTreeNews() //实盘待删除
	}
	toUpBitListBnExecute.GetExecute().StartTrig(s.symbolIndex, s.pScale, s.qScale, s.StMeta)
	// 开启第0秒协程
	limit := decimal.New(int64(priceLimit_U10), -bnConst.PScale_10).Truncate(s.pScale)
	toUpBitListBnExecute.GetExecute().PlacePostOnlyOrder(limit)
	//开启二次校验等待循环
	go func() {
		time.Sleep(2 * time.Second)
		if toUpBitListDataAfter.HasTreeNews.Load() {
			return
		}
		toUpBitListBnExecute.GetExecute().ReceiveStop(toUpBitListBnExecute.StopByTreeNews)
		toUpBitListDataStatic.SendToUpBitMsg("TreeNews未确认", map[string]string{
			"symbol": symbolName,
			"op":     "TreeNews未确认",
		})
	}()
	// 开启抽奖协程
	toUpBitListBnExecute.GetExecute().TryBuyLoop(20)
	// 获取止盈止损参数
	s.calParam()
	toUpBitListDataStatic.DyLog.GetLog().Infof("%s->[%s]价格触发,最新价格: %d,涨幅: %f%%,事件时间:%d", trigFlag, symbolName, priceTrig_8, riseValue*100, eventTs)
}
