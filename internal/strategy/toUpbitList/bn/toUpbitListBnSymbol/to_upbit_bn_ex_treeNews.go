package toUpbitListBnSymbol

import (
	"time"

	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/strategy/newsDrive/driverDefine"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
)

func (s *Single) checkTreeNews() {
	//开启二次校验等待循环
	go func() {
		time.Sleep(2 * time.Second)
		if s.hasTreeNews {
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
	s.hasTreeNews = true
	toUpBitDataStatic.SendToUpBitMsg("TreeNews确认", map[string]string{
		"symbol": s.StMeta.SymbolName,
		"op":     "TreeNews确认",
	})

	symbolName := s.StMeta.SymbolName
	s.TrigExType = exType
	switch exType {
	case exchangeEnum.UPBIT:
		s.tryBuyLoopUpBitKrw(20)
		toUpBitDataStatic.SendToUpBitMsg("TreeNews确认", map[string]string{"symbol": symbolName, "op": "upbit_TreeNews确认"})

	case exchangeEnum.BITHUMB:
		toUpBitDataStatic.SendToUpBitMsg("TreeNews确认", map[string]string{"symbol": symbolName, "op": "bithumb_TreeNews确认"})

	case exchangeEnum.BINANCE:
		safex.SafeGo("bnSpot_initBuyOpen", func() {
			s.initBnSpotBuyOpen()
			s.buyLoop()
		})
		toUpBitDataStatic.SendToUpBitMsg("TreeNews确认", map[string]string{"symbol": symbolName, "op": "binance_TreeNews确认"})
	}
}

func (s *Single) ReceiveNoTreeNews() {
	s.hasTreeNews = false
	s.receiveStop(driverDefine.StopByTreeNews)
	toUpBitDataStatic.SendToUpBitMsg("TreeNews未确认", map[string]string{
		"symbol": s.StMeta.SymbolName,
		"op":     "TreeNews未确认",
	})
}
