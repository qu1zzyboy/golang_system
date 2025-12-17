package toUpbitListBnSymbol

import (
	"fmt"

	"upbitBnServer/internal/quant/exchanges/binance/bnConst"
	"upbitBnServer/internal/quant/market/symbolInfo/coinMesh"
	"upbitBnServer/internal/strategy/newsDrive/driverDefine"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitBnMode"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"

	"github.com/shopspring/decimal"
)

//to do

func (s *Single) onOrderPriceCheck(tradeTs int64, priceU64_8 uint64) {
	// minBId>=0.95*markPrice
	if float64(s.minPriceAfterMp) >= toUpBitDataStatic.PriceRiceTrig*float64(s.markPrice_8) {
		s.IntoExecuteNoCheck(tradeTs, "preOrder", priceU64_8)
	} else {
		toUpBitDataStatic.SendToUpBitMsg("成交但不满足上市check消息失败", map[string]string{
			"msg":  fmt.Sprintf("成交但不满足上市check,成交价:%d", priceU64_8),
			"bn品种": s.StMeta.SymbolName,
			"上涨幅度": fmt.Sprintf("%.2f%%", s.lastRiseValue*100),
		})
	}
}

func (s *Single) IntoExecuteNoCheck(eventTs int64, trigFlag string, priceTrig_8 uint64) {
	s.hasTreeNews = toUpbitBnMode.Mode.GetTreeNewsFlag()
	toUpBitListDataAfter.Trig(s.symbolIndex)
	s.startTrig()
	limit := decimal.New(int64(s.priceMaxBuy_10), -bnConst.PScale_10).Truncate(s.pScale)
	// debug版默认为true,不会收到消息也不会退出
	s.checkTreeNews()
	s.PlacePostOnlyOrder(limit)
	s.TryBuyLoop(20)
	// 获取止盈止损参数
	s.calParam()
	toUpBitDataStatic.DyLog.GetLog().Infof("%s->[%s]价格触发,最新价格: %d,涨幅: %f%%,事件时间:%d",
		trigFlag, s.StMeta.SymbolName, priceTrig_8, s.lastRiseValue*100, eventTs)
}

func (s *Single) intoExecuteByMsg() {
	s.hasTreeNews = true
	toUpBitListDataAfter.Trig(s.symbolIndex)
	s.startTrig()
	s.TryBuyLoop(20)
	// 获取止盈止损参数
	s.calParam()
	toUpBitDataStatic.DyLog.GetLog().Infof("treeNews->[%s]触发,涨幅: %f%%", s.StMeta.SymbolName, s.lastRiseValue*100)
}

func (s *Single) calParam() {
	symbolName := s.StMeta.SymbolName
	//获取流通量
	mesh, ok := coinMesh.GetManager().Get(s.StMeta.TradeId)
	if !ok {
		toUpBitDataStatic.DyLog.GetLog().Errorf("coin mesh [%s] not found for tradeId: %d", symbolName, s.StMeta.TradeId)
		s.receiveStop(driverDefine.StopByGetCmcFailure)
		toUpBitDataStatic.SendToUpBitMsg("获取cmc_id失败", map[string]string{
			"symbol": symbolName,
			"op":     "获取cmc_id失败",
		})
		return
	}
	// 2min之前的市值
	last2MinCloseF64 := float64(s.last2MinClose_8) / 1e8
	cap2Min := mesh.SupplyNow * last2MinCloseF64
	//计算止盈止损参数
	gainPct, twapSec, err := toUpbitBnMode.Mode.GetTakeProfitParam(mesh.IsMeMe, s.symbolIndex, cap2Min/1_000_000)
	if err != nil {
		toUpBitDataStatic.DyLog.GetLog().Errorf("coin mesh [%s] 获取止盈止损失败: %v", symbolName, err)
		s.receiveStop(driverDefine.StopByGetRemoteFailure)
		toUpBitDataStatic.SendToUpBitMsg("获取止盈止损失败", map[string]string{
			"symbol": symbolName,
			"op":     "获取止盈止损失败",
		})
		return
	}
	// 返回值格式 15.5 30
	toUpBitDataStatic.DyLog.GetLog().Infof("远程参数:%t,市值:%f,%s,远程响应:[%f,%f]", mesh.IsMeMe, cap2Min/1_000_000, symbolName, gainPct, twapSec)
	s.setExecuteParam(last2MinCloseF64*(1+0.01*(gainPct)), twapSec)
}
