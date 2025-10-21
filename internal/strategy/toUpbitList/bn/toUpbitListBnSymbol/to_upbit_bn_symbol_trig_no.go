package toUpbitListBnSymbol

import (
	"fmt"
	"time"

	"github.com/hhh500/quantGoInfra/quant/exchanges/binance/bnConst"
	"github.com/hhh500/upbitBnServer/internal/quant/market/symbolInfo/coinMesh"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/bn/toUpBitListBnExecute"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"github.com/shopspring/decimal"
)

func (s *Single) IntoExecuteNoCheck(eventTs int64, trigFlag string, riseValue float64, priceTrig_8 uint64) {
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
		toUpBitListDataAfter.UpdateTreeNewsFlag(s.symbolIndex)           //实盘待删除
		toUpBitListBnExecute.GetExecute().ReceiveTreeNews(s.symbolIndex) //实盘待删除
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

func (s *Single) calParam() {
	symbolName := s.StMeta.SymbolName
	//获取流通量
	mesh, ok := coinMesh.GetManager().Get(s.StMeta.TradeId)
	if !ok {
		toUpBitListDataStatic.DyLog.GetLog().Errorf("coin mesh [%s] not found for tradeId: %d", symbolName, s.StMeta.TradeId)
		toUpBitListBnExecute.GetExecute().ReceiveStop(toUpBitListBnExecute.StopByGetCmcFailure)
		toUpBitListDataStatic.SendToUpBitMsg("获取cmc_id失败", map[string]string{
			"symbol": symbolName,
			"op":     "获取cmc_id失败",
		})
		return
	}
	// 2min之前的市值
	last2MinCloseF64 := float64(s.last2MinClose_8) / 1e8
	cap2Min := mesh.SupplyNow * last2MinCloseF64
	//计算止盈止损参数
	gainPct, twapSec, err := GetParam(mesh.IsMeMe, cap2Min/1_000_000, symbolName)
	if err != nil {
		toUpBitListDataStatic.DyLog.GetLog().Errorf("coin mesh [%s] 获取止盈止损失败: %v", symbolName, err)
		toUpBitListBnExecute.GetExecute().ReceiveStop(toUpBitListBnExecute.StopByGetRemoteFailure)
		toUpBitListDataStatic.SendToUpBitMsg("获取止盈止损失败", map[string]string{
			"symbol": symbolName,
			"op":     "获取止盈止损失败",
		})
		return
	}
	// 返回值格式 15.5 30
	toUpBitListDataStatic.DyLog.GetLog().Infof("远程参数:%t,市值:%f,%s,远程响应:[%f,%f]", mesh.IsMeMe, cap2Min/1_000_000, symbolName, gainPct, twapSec)
	toUpBitListBnExecute.GetExecute().SetExecuteParam(last2MinCloseF64*(1+0.01*gainPct), twapSec)
}
