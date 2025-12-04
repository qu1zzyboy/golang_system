package treenews

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestManualTreeNews(t *testing.T) {
    RegisterHandler(func(_ context.Context, evt Event) {
        t.Logf("exchange=%s symbols=%v id=%s", evt.Exchange, evt.Symbols, evt.ID)
    })

    cfg := defaultConfig()
    cfg.Enabled = false // 避免真正连 WS
    svc := NewService(cfg)
    svc.dedup = newIDSet(1024)

    raw := `{
        "_id":"demo-binance-001",
        "source":"Binance EN",
        "title":"Binance Will List Lorenzo Protocol (BANK) and Meteora (MET) with Seed Tag Applied",
        "en":"Binance Will List Lorenzo Protocol (BANK) and Meteora (MET) with Seed Tag Applied",
        "time": %d
    }`
    msg := queuedMessage{
        data: []byte(fmt.Sprintf(raw, time.Now().UnixMilli())),
        recv: time.Now().UTC(),
    }

    svc.processMessage(context.Background(), msg)
}
