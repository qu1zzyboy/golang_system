# ReceiveTreeNews 行为与扩展方向

## 现状
- `tree_news_bridge.go` 中注册的 handler 接收到 Tree News 事件后，定位触发的 `symbolIndex`，并调用 `Single.ReceiveTreeNews()` 或 `ReceiveNoTreeNews()`。
- `ReceiveTreeNews()` 仅更新 `hasTreeNews` 标记，并通过 `SendToUpBitMsg` 投递一条“TreeNews确认”通知；`ReceiveNoTreeNews()` 则清空标记、触发 `StopByTreeNews` 且发送“TreeNews未确认”。
- handler 目前不区分交易所，上层策略仍依赖预挂单成交时的默认涨跌幅 / TWAP 计算。

## 后续可行方向
1. **区分交易所**：利用 `treenews.Event.Exchange`，在 handler 中区分 Upbit / Binance（及未来更多交易所）并传递不同的参数。
2. **带参数的回调**：将 `Single.ReceiveTreeNews()` 扩展为接受“交易所 + 新的 gain/TWAP”或新增专用方法，例如 `ReceiveTreeNewsWithParam(exchange string, gain float64, twap float64)`。
3. **参数来源**：新的 gain/TWAP 可直接调用 `toUpbitParam` 中的 `ExpectedSplitGainAndTwapDurationWithExchange()`，以市值、恐惧贪婪指数、BTC 走势、meme 标签等维度生成各交易所专属配置。
4. **通知增强**：在 Telegram 信息中附带交易所与参数信息，便于监控和回溯 Tree News 对应的动作。

此文档用于同步当前约束和未来改造路径，等交易所区分逻辑落地后再迭代实现细节。
