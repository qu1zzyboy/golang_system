package bnDriveSymbolArr

import (
	"context"
	"strings"
	"upbitBnServer/internal/strategy/newsDrive/common/driverStatic"

	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
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
		symbolIndexTrue, ok := driverStatic.SymbolIndex.Load(symbolName)
		if !ok {
			driverStatic.DyLog.GetLog().Errorf("%s treeNews品种不在品种池内", symbolName)
			continue
		}
		driverStatic.DyLog.GetLog().Infof("received tree news: symbol=%s id=%s", symbolName, evt.ID)
		// 触发品种和TreeNews品种一致
		if symbolIndexTrue == toUpBitListDataAfter.TrigSymbolIndex {
			GetSymbolObj(symbolIndexTrue).ReceiveTreeNews()
		} else {
			GetSymbolObj(toUpBitListDataAfter.TrigSymbolIndex).ReceiveNoTreeNews()
		}
		return
	}
}
