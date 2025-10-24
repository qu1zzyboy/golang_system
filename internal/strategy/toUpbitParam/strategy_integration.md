# Params 参数服务接入指南

本文聚焦于 `internal/strategy/params` 模块，说明如何启动参数服务并获取 “涨幅（GainPct）”“TWAP 时间（TwapSec）” 等指标，方便后续流程调用。

---

## 1. 模块概览
- **功能**：根据市值、恐惧贪婪指数、BTC 收益、OI 数据等要素，计算策略所需的目标涨幅（`GainPct`）与执行周期（`TwapSec`）。
- **数据来源**：
  - Binance Futures REST 接口：拉取 BTC K 线，计算 24h/7d 收益；
  - Fear & Greed API：周期性获取指数；
  - Redis（`BN_OPEN_INTEREST` 哈希）：获取 OI 数据；
  - 本地配置：市值桶、OI 修正、止盈/止损等。
- **启动方式**：通过 `bootx` 统一管理，既可以直接调用，也可以嵌入服务进程。

---

## 2. 依赖与配置

### 2.1 依赖模块
调用参数服务前，需注册以下 boot：

```go
manager := bootx.GetManager()
manager.Register(conf.NewBoot())            // 解析 config_main.yaml
manager.Register(safex.NewBoot())           // 协程安全工具（SafeGo 等）
manager.Register(redisConfig.NewBoot())     // Redis 客户端
manager.Register(params.NewBoot())          // 参数服务
```

如需输出通知/日志，可额外注册 `notify.NewBoot(...)` 等模块。

### 2.2 配置文件
`config/config_main.yaml` 中需提供 Redis 地址、Tree News 开关等信息；参数服务无独立配置段落，如需调整数据源或限值，请在代码或环境变量中覆写。常见做法：

```bash
export CONFIG=config/config_main.yaml
go run ./test/compute-live -cap 50 -symbol BTCUSDT
```

---

## 3. 启动与关闭

```go
ctx := context.Background()
manager.StartAll(ctx)
defer manager.StopAll(ctx)  // 进程退出前调用
```

`StartAll` 会处理依赖关系并保证各模块仅启动一次。测试时如需设置超时，可使用 `context.WithTimeout`。

---

## 4. 计算接口：`Service.Compute`

### 4.1 请求结构
```go
type ComputeRequest struct {
    MarketCapM float64 // 市值（百万美元）
    IsMeme     bool    // 是否 Meme/特殊类别
    SymbolName string  // Symbol（用于 OI/事件匹配，可为空）
}
```

### 4.2 响应结构
```go
type ComputeResponse struct {
    GainPct float64      // 建议涨幅（百分比）
    TwapSec float64      // 建议 TWAP 时间（秒）
    Diag    Diagnostics  // 诊断信息（用于观测和排错）
}
```

`Diagnostics` 包含基础收益、OI 修正、BTC 指标、FGI、事件 staleness 等信息，方便排查数据是否过期。

### 4.3 使用示例
```go
svc := params.GetService()
resp, err := svc.Compute(context.Background(), params.ComputeRequest{
    MarketCapM: 80,          // 80M USDT 市值
    IsMeme:     false,
    SymbolName: "BTCUSDT",
})
if err != nil {
    // 错误处理
    log.Fatal(err)
}
fmt.Printf("GainPct=%.2f%%, TwapSec=%.1f\n", resp.GainPct, resp.TwapSec)
```

---

## 5. 诊断字段说明

`resp.Diag` 常用字段：

| 字段 | 含义 |
|------|------|
| `GainBase` / `TwapBase` | 市值分桶计算的基础值 |
| `GainOIAdd` / `TwapOIAdd` | OI 修正贡献 |
| `GainFinal` / `TwapFinal` | 裁剪后的结果，与 `GainPct`/`TwapSec` 一致 |
| `FGI` | 恐惧贪婪指数 |
| `BTC1D` / `BTC7D` | BTC 收益（百分比） |
| `OI` | 对应 symbol 的 OI |
| `StalenessSeconds` | BTC 数据滞后时间（秒） |

便于后续写监控或报警，例如当 `StalenessSeconds` 过大时可提示 BTC 指标过期。

---

## 6. 测试与验证

### 6.1 本地拉取真实数据
`test/compute-live/main.go` 会启动所需 boot 并调用 `Compute`：

```bash
go run ./test/compute-live -cap 50 -symbol BTCUSDT
```

输出类似：
```
Compute(BTCUSDT, cap=50.00, meme=false) => gain=25.30 twap=48.0 diag={...}
```

### 6.2 纯单元测试（Mock）
使用 `params.NewWithProviders` 注入 stub 实现，避开外部依赖：

```go
svc := params.NewWithProviders(
    params.Config{},
    &stubBTC{snap: params.BTCSnapshot{BTC1D: 1.2, BTC7D: 3.4}},
    &stubFGI{val: 65},
    &stubOI{rec: map[string]params.OIRecord{"AAAUSDT": {...}}},
)
_ = svc.Start(context.Background())
resp, _ := svc.Compute(ctx, params.ComputeRequest{MarketCapM: 50, SymbolName: "AAAUSDT"})
```

这类测试可覆盖特定分桶、OI 缺失、FGI 为空等场景。

---

## 7. 下游消费建议

1. **策略执行**：将 `GainPct`、`TwapSec` 作为策略阈值或下单参数；`Diag` 可用于判定是否符合准入条件（例如 OI 过旧、FGI 极值等）。
2. **监控报警**：
   - 基于 `Diag.StalenessSeconds` 监测行情/API 异常；
   - 当 `GainPct`/`TwapSec` 异常变化时，结合 `Diag.GainBase/GainOIAdd` 判断是配置变化还是数据异常。
3. **缓存与节流**：
   - 若多个下游服务同时调用，可在外部实现缓存层，或约束调用频率（例如按分钟聚合，避免重复请求）。

---

## 8. 常见问题

1. **Compute 返回延迟过大/超时**  
   - 检查 BTC/Fear&Greed API 是否可访问；必要时调整 `timeout` 或增加重试；
   - Redis 连接失败会导致 OI 信息缺失，可在 `Diag.OI` 查看。

2. **结果异常跳变**  
   - 首先在 `Diag` 中查看 `GainBase`/`GainOIAdd` 是否正常；
   - 检查市值桶配置、OI 数据是否更新。

3. **离线环境测试**  
   - 建议通过 stub 注入的方式（`NewWithProviders`）模拟数据；
   - 或将 BTC/FGI/OI 抓取脚本提前落盘，再在测试代码中加载。

---

通过上述步骤，可以快速集成参数服务，为不同业务场景（策略执行、监控报警等）提供统一的涨幅/TWAP 参考值。若后续需要扩展更多指标或模型，只需在 `ComputeResponse` 中追加字段即可。 
