package toUpbitListBnSymbol

import (
	"context"
	"time"

	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/quant/market/symbolInfo/coinMesh"
	"upbitBnServer/internal/strategy/newsDrive/driverDefine"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitBnMode"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
)

func (s *Single) checkTreeNews() {
	//开启二次校验等待循环
	go func() {
		time.Sleep(2 * time.Second)
		if s.hasTreeNews.Load() {
			return
		}
		s.receiveStop(driverDefine.StopByTreeNews)
		toUpBitDataStatic.SendToUpBitMsg("TreeNews未确认", map[string]string{
			"symbol": s.StMeta.SymbolName,
			"op":     "TreeNews未确认",
		})
	}()
}

func (s *Single) ReceiveTreeNews(exType exchangeEnum.ExchangeType) {
	toUpBitDataStatic.DyLog.GetLog().Info("--------------------TreeNews确认---------------------")
	s.hasTreeNews.Store(true)
	toUpBitDataStatic.SendToUpBitMsg("TreeNews确认", map[string]string{"symbol": s.StMeta.SymbolName, "op": "TreeNews确认"})

	// 获取止盈止损参数
	s.calParam(exType)

	symbolName := s.StMeta.SymbolName
	s.TrigExType = exType
	switch exType {
	case exchangeEnum.UPBIT:
		s.tryBuyLoopUpBitKrw(20)
		toUpBitDataStatic.SendToUpBitMsg("TreeNews确认", map[string]string{"symbol": symbolName, "op": "upbit_TreeNews确认"})

	case exchangeEnum.BITHUMB:
		toUpBitDataStatic.SendToUpBitMsg("TreeNews确认", map[string]string{"symbol": symbolName, "op": "bithumb_TreeNews确认"})

	case exchangeEnum.BINANCE:
		s.bnSpotCtxStop, s.bnSpotCancel = context.WithCancel(context.Background())
		s.bnBeginTwapBuy = 1.19 * s.TrigMartPrice
		s.tryBuyLoopBnSpot(20)
		toUpBitDataStatic.SendToUpBitMsg("TreeNews确认", map[string]string{"symbol": symbolName, "op": "binance_TreeNews确认"})
	}
}

func (s *Single) ReceiveNoTreeNews() {
	s.hasTreeNews.Store(false)
	s.receiveStop(driverDefine.StopByTreeNews)
	toUpBitDataStatic.SendToUpBitMsg("TreeNews未确认", map[string]string{"symbol": s.StMeta.SymbolName, "op": "TreeNews未确认"})
}

func (s *Single) calParam(exType exchangeEnum.ExchangeType) {
	symbolName := s.StMeta.SymbolName
	//获取流通量
	mesh, ok := coinMesh.GetManager().Get(s.StMeta.TradeId)
	if !ok {
		toUpBitDataStatic.DyLog.GetLog().Errorf("coin mesh [%s] not found for tradeId: %d", symbolName, s.StMeta.TradeId)
		s.receiveStop(driverDefine.StopByGetCmcFailure)
		toUpBitDataStatic.SendToUpBitMsg("获取cmc_id失败", map[string]string{"symbol": symbolName, "op": "获取cmc_id失败"})
		return
	}
	// 2min之前的市值
	last2MinCloseF64 := float64(s.last2MinClose_8) / 1e8
	cap2Min := mesh.SupplyNow * last2MinCloseF64
	//计算止盈止损参数
	gainPct, twapSec, err := toUpbitBnMode.Mode.GetTakeProfitParam(exType, mesh.IsMeMe, s.SymbolIndex, cap2Min/1_000_000)
	if err != nil {
		toUpBitDataStatic.DyLog.GetLog().Errorf("coin mesh [%s] 获取止盈止损失败: %v", symbolName, err)
		s.receiveStop(driverDefine.StopByGetRemoteFailure)
		toUpBitDataStatic.SendToUpBitMsg("获取止盈止损失败", map[string]string{"symbol": symbolName, "op": "获取止盈止损失败"})
		return
	}
	// 返回值格式 15.5 30
	toUpBitDataStatic.DyLog.GetLog().Infof("远程参数:%t,市值:%f,%s,远程响应:[%f,%f]", mesh.IsMeMe, cap2Min/1_000_000, symbolName, gainPct, twapSec)
	s.setExecuteParam(last2MinCloseF64*(1+0.01*(gainPct)), twapSec)
}
