package toUpBitListBnExecute

import (
	"context"
	"strings"

	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"github.com/hhh500/upbitBnServer/internal/strategy/treenews"
)

func init() {
	treenews.RegisterHandler(treeNewsHandler)
}

func treeNewsHandler(_ context.Context, evt treenews.Event) {
	if len(evt.Symbols) == 0 {
		return
	}
	for _, raw := range evt.Symbols {
		symbol := strings.ToUpper(raw)
		if !strings.HasSuffix(symbol, "USDT") {
			symbol = symbol + "USDT"
		}
		index, ok := toUpBitListDataStatic.SymbolIndex.Load(symbol)
		if !ok {
			continue
		}
		toUpBitListDataStatic.DyLog.GetLog().Infof("received tree news: symbol=%s id=%s", symbol, evt.ID)
		GetExecute().ReceiveTreeNews(index)
		toUpBitListDataAfter.UpdateTreeNewsFlag(index)
		return
	}
}
