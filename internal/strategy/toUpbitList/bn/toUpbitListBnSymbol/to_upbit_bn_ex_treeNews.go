package toUpbitListBnSymbol

import (
	"time"

	exchangeEnum "upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitDefine"
)

func (s *Single) checkTreeNews() {
	//开启二次校验等待循环
	go func() {
		time.Sleep(2 * time.Second)
		if s.hasTreeNews {
			return
		}
		s.receiveStop(toUpbitDefine.StopByTreeNews)
		toUpBitListDataStatic.SendToUpBitMsg("TreeNews未确认", map[string]string{
			"symbol": s.StMeta.SymbolName,
			"op":     "TreeNews未确认",
		})
	}()
}

func (s *Single) ReceiveTreeNews() {
	s.ReceiveTreeNewsWithExchange(exchangeEnum.UPBIT)
}

func (s *Single) ReceiveTreeNewsWithExchange(exType exchangeEnum.ExchangeType) {
	s.treeNewsExchangeType = normalizeExchangeType(exType)
	s.hasTreeNews = true
	toUpBitListDataStatic.SendToUpBitMsg("TreeNews确认", map[string]string{
		"symbol": s.StMeta.SymbolName,
		"op":     "TreeNews确认",
	})
}

func (s *Single) ReceiveNoTreeNews() {
	s.hasTreeNews = false
	s.treeNewsExchangeType = exchangeEnum.UPBIT
	s.receiveStop(toUpbitDefine.StopByTreeNews)
	toUpBitListDataStatic.SendToUpBitMsg("TreeNews未确认", map[string]string{
		"symbol": s.StMeta.SymbolName,
		"op":     "TreeNews未确认",
	})
}

func normalizeExchangeType(ex exchangeEnum.ExchangeType) exchangeEnum.ExchangeType {
	switch ex {
	case exchangeEnum.BINANCE:
		return exchangeEnum.BINANCE
	case exchangeEnum.UPBIT:
		return exchangeEnum.UPBIT
	default:
		return exchangeEnum.UPBIT
	}
}
