package toUpbitListBnSymbolArr

import (
	"context"
	"strings"

	exchangeEnum "upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"upbitBnServer/internal/strategy/treenews"
)

func init() {
	treenews.RegisterHandler(treeNewsHandler)
}

func treeNewsHandler(_ context.Context, evt treenews.Event) {
	if len(evt.Symbols) == 0 {
		return
	}
	for _, raw := range evt.Symbols {
		symbolName := strings.ToUpper(raw)
		if !strings.HasSuffix(symbolName, "USDT") {
			symbolName = symbolName + "USDT"
		}
		symbolIndexTrue, ok := toUpBitListDataStatic.SymbolIndex.Load(symbolName)
		if !ok {
			toUpBitListDataStatic.DyLog.GetLog().Errorf("%s treeNews品种不在品种池内", symbolName)
			continue
		}
		exType := normalizeExchangeType(evt.ExchangeType)
		exchange := evt.Exchange
		if exchange == "" {
			exchange = exType.String()
			if exchange == "ERROR" || exchange == "" {
				exchange = "unknown"
			}
		}
		toUpBitListDataStatic.DyLog.GetLog().Infof("received tree news: exchange=%s exchange_type=%s symbol=%s id=%s", exchange, exType.String(), symbolName, evt.ID)
		// 触发品种和TreeNews品种一致
		if symbolIndexTrue == toUpBitListDataAfter.TrigSymbolIndex {
			sym := GetSymbolObj(symbolIndexTrue)
			if exType == exchangeEnum.BINANCE {
				sym.ReceiveTreeNewsWithExchange(exchangeEnum.BINANCE)
			} else {
				sym.ReceiveTreeNews()
			}
		} else {
			GetSymbolObj(toUpBitListDataAfter.TrigSymbolIndex).ReceiveNoTreeNews()
		}
		return
	}
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
