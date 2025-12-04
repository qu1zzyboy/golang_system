# ReceiveTreeNews 行为与扩展方向

## 现状
- `tree_news_bridge.go` 解析 `treenews.Event.ExchangeType`，区分 Upbit / Binance，并分别调用 `Single.ReceiveTreeNews()`（默认 Upbit）或 `ReceiveTreeNewsWithExchange(exchangeEnum.BINANCE)`。
- `ReceiveTreeNewsWithExchange()` 会记录交易所类型，沿用原有的 Telegram 通知流程；`ReceiveNoTreeNews()` 清理 `hasTreeNews` 并将交易所重置为 Upbit。
- `calParam()` 在计算止盈/平仓参数时，会根据 `hasTreeNews` 与 `treeNewsExchangeType` 选择传递给 `toUpbitParam` 的交易所枚举，从而通过 `ExpectedSplitGainAndTwapDurationWithExchange()` 获得 Upbit/ Binance 各自的 gain/TWAP。

## 后续可行方向
1. **更多交易所**：当 Tree News 接入第三家交易所时，只需在 handler 中追加分支，并扩展 `toUpbitParam` 的分桶配置即可。
2. **显式参数下发**：若需要在 Tree News 到达即透出具体的 gain/TWAP，可在 handler 计算完后附带在 Telegram/日志中，便于人工确认。
3. **撤销逻辑优化**：目前 `ReceiveNoTreeNews()` 直接回落到默认参数，可考虑在 Binance 撤销时保留最近一次参数用于审计。

此文档用于同步已实现的行为，以及未来扩展时需要注意的方向。
