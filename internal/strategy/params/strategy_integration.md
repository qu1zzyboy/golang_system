# Strategy 模块接入指南

本文面向需要复用 `internal/strategy` 下功能（策略参数服务、Tree News 事件）的其它模块，说明最小集成步骤。

## 目录

1. [核心能力概览](#核心能力概览)  
2. [快速接入流程](#快速接入流程)  
   1. [依赖注入](#依赖注入)  
   2. [启动与停止](#启动与停止)  
   3. [调用参数服务](#调用参数服务)  
   4. [订阅 Tree News 事件（可选）](#订阅-tree-news-事件可选)  
3. [本地测试示例](#本地测试示例)  
4. [Mock 建议](#mock-建议)  
5. [Tree News 独立验证](#tree-news-独立验证)

---

## 核心能力概览

模块 | 功能 | 产出
---|---|---
`internal/strategy/params` | 根据市值、情绪、BTC 走势与 OI 计算止盈与 TWAP | `Compute` 返回 `GainPct`、`TwapSec` 及诊断信息
`internal/strategy/treenews` | 监听 Tree News WebSocket，筛选 Upbit KRW 相关事件 | 回调 `treenews.Event`，驱动策略执行

两者均实现 `bootx.Bootable`，可通过 `bootx` 统一启动。

---

## 快速接入流程

### 依赖注入

```go
manager := bootx.GetManager()
manager.Register(conf.NewBoot())                                  // 读取配置
manager.Register(notify.NewBoot(notifyTg.GetTg()))                // 观测通知
manager.Register(safex.NewBoot())                                 // 协程安全
manager.Register(redisConfig.NewBoot())                           // Redis 客户端
manager.Register(params.NewBoot())                                // 参数服务
manager.Register(treenews.NewBoot())                              // Tree News 监听（如需要）
```

> 说明：各模块依赖关系（`DependsOn`）已在定义中声明，上述注册顺序可保持一致。

### 启动与停止

```go
ctx := context.Background()
manager.StartAll(ctx)
defer manager.StopAll(ctx)
```

`StartAll` 会递归处理依赖并确保每个组件仅启动一次。若需在 CLI/测试中限制启动耗时，请使用 `context.WithTimeout`。

### 调用参数服务

```go
svc := params.GetService()
resp, err := svc.Compute(ctx, params.ComputeRequest{
    MarketCapM: 50,
    IsMeme:     false,
    SymbolName: "BTCUSDT",
})
if err != nil {
    // 处理错误
}
fmt.Printf("gain=%.2f twap=%.2f\n", resp.GainPct, resp.TwapSec)
```

`Compute` 同步返回，内部会使用最新缓存；若指标滞后或缺失，诊断信息（`resp.Diag`）可帮助定位。

#### 可配置项

策略默认从 `conf.RedisCfg`、线上 REST 接口获取行情。若需要静态配置，请在 `Start` 前调用：

```go
svc := params.GetService()
_ = svc.SetConfig(params.Config{ /* 自定义参数 */ })
_ = svc.SetProviders(btcMock, fgiMock, oiMock) // 仅测试使用
```

### 订阅 Tree News 事件（可选）

Tree News 模块启动后，即可注册回调：

```go
treenews.RegisterHandler(func(ctx context.Context, evt treenews.Event) {
    fmt.Printf("Tree News %s symbols=%v\n", evt.ID, evt.Symbols)
})
```

策略默认桥接逻辑在 `internal/strategy/toUpbitList/bn/toUpBitListBnExecute/tree_news_bridge.go`，如需自定义行为，可额外注册 handler。配置项通过环境变量控制，常用变量：

环境变量 | 默认值 | 含义
---|---|---
`TREE_NEWS_ENABLED` | `false` | 是否启用 Tree News
`TREE_NEWS_API_KEY` | 示例 key | 登录凭证
`TREE_NEWS_URL` | `wss://news.treeofalpha.com/ws` | 目标 WebSocket 地址

---

## 本地测试示例

仓库已提供两个示例：

1. `test/compute-demo`: 使用 stub 注入，验证公式正确性。  
2. `test/compute-live`: 启动上述组件，真实拉取行情 & Redis 数据。运行方式：
   ```bash
   go run ./test/compute-live -cap 80 -symbol BTCUSDT
   ```

输出包含 gain/twap 结果和诊断字段。

---

## Mock 建议

### 参数服务

实现接口替换即可：

```go
type stubBTC struct{ snap params.BTCSnapshot }
func (s *stubBTC) Start(context.Context) error { return nil }
func (s *stubBTC) Stop(context.Context) error  { return nil }
func (s *stubBTC) Snapshot() params.BTCSnapshot { return s.snap }

svc := params.NewWithProviders(
    params.Config{},
    &stubBTC{snap: params.BTCSnapshot{BTC1D: 1.2, BTC7D: 3.4}},
    &stubFGI{val: 55},
    &stubOI{rec: map[string]params.OIRecord{ /* ... */ }},
)
```

也可以使用 `SetProviderFactories` 在 `Start` 前注入工厂函数。

### Tree News

1. 使用 `net/http/httptest` + gorilla/websocket 启动本地 mock 服务器，推送测试 JSON。  
2. 设置 `TREE_NEWS_URL` 指向 mock 地址，运行 `treenews.NewBoot()`。  
3. 在测试中通过 `RegisterHandler` 捕获事件并断言。

---

## Tree News 独立验证

若仅需确认 WebSocket 链接与过滤逻辑，可单独启动 Tree News 模块：

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/hhh500/quantGoInfra/conf"
	"github.com/hhh500/quantGoInfra/infra/bootx"
	"github.com/hhh500/quantGoInfra/infra/observe/notify"
	"github.com/hhh500/quantGoInfra/infra/observe/notify/notifyTg"
	"github.com/hhh500/quantGoInfra/infra/safex"
	"github.com/hhh500/upbitBnServer/internal/strategy/treenews"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	manager := bootx.GetManager()
	manager.Register(conf.NewBoot())
	manager.Register(notify.NewBoot(notifyTg.GetTg()))
	manager.Register(safex.NewBoot())
	manager.Register(treenews.NewBoot())

	treenews.RegisterHandler(func(_ context.Context, evt treenews.Event) {
		fmt.Printf("Tree News: id=%s symbols=%v\n", evt.ID, evt.Symbols)
	})

	manager.StartAll(ctx)
	defer manager.StopAll(context.Background())

	<-ctx.Done()
}
```

运行前设置：

- `TREE_NEWS_ENABLED=1`
- `TREE_NEWS_API_KEY=<有效 key>`
- `TREE_NEWS_URL`（若需指定本地 mock）

日志 (`logs/` 目录中包含 “tree news”) 或 handler 输出可用来判断是否成功连接、登陆以及过滤消息。

---

通过以上步骤，可在不依赖完整 gRPC 服务的情况下，将策略模块嵌入其它子系统，实现参数获取与 Tree News 事件驱动。若有更多需求（如并发控制、诊断指标订阅），可进一步参考 `internal/strategy/params` 与 `internal/strategy/treenews` 源码。
