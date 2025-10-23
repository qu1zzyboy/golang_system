## Tree News 模块说明

Tree News 模块负责监听 TreeOfAlpha 的 WebSocket，筛选与 Upbit KRW 上市相关的新闻事件，并通过 channel 推送给业务协程。本文档记录配置项、运行流程以及下游消费建议，方便将消息衔接到后续流程（例如策略执行、告警系统等）。

---

### 一、模块结构

1. **配置读取**
   - `config/config_main.yaml` 中的 `treeNews` 段落承载全部运行参数；
   - `libs/quantGoInfra/conf/conf_boot.go` 启动时读取上述配置并写入 `conf.TreeNewsCfg`；
   - 模块支持环境变量覆写（例如 `TREE_NEWS_ENABLED`、`TREE_NEWS_API_KEY` 等，配置优先级为：环境变量 > YAML > 内置默认值）。

2. **核心组件**
   - `Service` 是单例：管理 WebSocket 连接、监听协程、消息去重；
   - `readLoop`：只负责把 raw 消息读入有界 channel（`outQueue`），实现背压；
   - `mergerLoop`：顺序消费 `outQueue`，完成 JSON 解析、延迟计算、过滤以及业务回调；
   - `workerState`：记录 ping/pong RTT、连续异常次数，支持高延迟自动重连；
   - `RegisterHandler(fn)`：注册业务回调（可多播），用于后续流程消费 Tree News 事件。

3. **日志位置**
   - 普通日志：`logs/<date>/tree_news_log/tree_news.log`；
   - 错误日志：`logs/<date>/tree_news_log/tree_news_error.log`；
   - 日志包含连接建立、心跳 RTT、raw 消息延迟、过滤后事件、超阈告警/重连等信息。

---

### 二、配置项说明（`config/config_main.yaml`）

```yaml
treeNews:
  enabled: true                   # 是否启用 Tree News
  apiKey: "xxx"                   # WebSocket 登录所需的 API Key
  url: "wss://news.treeofalpha.com/ws"
  workers: 2                      # 并行连接数
  pingInterval: "15s"             # 心跳发送间隔
  pingTimeout: "2s"               # ping 超时时间
  rollingReconnect: "1h"          # 定期重新建立连接的周期
  rollingJitter: "10m"            # 滚动重连的随机抖动
  dedupCapacity: 50000            # 消息去重缓存大小
  queueCapacity: 50000            # 读协程 → 业务协程的有界队列容量
  latencyWarnMs: 500              # 消息延迟告警阈值（毫秒）
  latencyWarnCount: 3             # 连续达到阈值的次数后触发重连
  rttWarnMs: 400                  # 心跳 RTT 告警阈值（毫秒）
  rttWarnCount: 3                 # 连续达到阈值的次数后触发重连
```

> **Tip**：若某个值需要临时调整，可设置环境变量，例如：
> ```bash
> export TREE_NEWS_LATENCY_WARN_MS=600
> export TREE_NEWS_ENABLED=1
> ```

---

### 三、事件数据结构

`treenews.Event`（推送给业务 handler）的主要字段：

| 字段 | 说明 |
|------|------|
| `ID` | Tree News 原始 `_id`（若缺失则使用 `raw-序号`） |
| `Symbols` | 通过 Upbit KRW 过滤后的交易对列表 |
| `Payload` | 原始 JSON 内容 |
| `ReceivedAt` | 客户端收到消息的时间戳（UTC） |
| `ServerMilli` | Tree News payload 中的服务器时间（毫秒） |
| `LatencyRawMS` | 原始延迟：`ReceivedAt - ServerMilli` |
| `LatencyMS` | 与 `LatencyRawMS` 当前保持一致，预留用于意图延时修正 |
| `RTTMS` | 最近一次 ping/pong RTT（毫秒） |

> 业务回调可直接使用 `Symbols` 做后续判断；如果要关联完整 payload，可读取 `Payload` 字段。

---

### 四、运行流程（简述）

1. 调用 `treenews.NewBoot()`（需依赖 `conf.NewBoot()`），执行 `Service.Start()`；
2. 每个 worker 建立 WS → 发送 login → 启动读协程、心跳协程、滚动重连；
3. `readLoop` 接收 raw 消息并推入 `outQueue`；
4. `mergerLoop` 顺序处理消息：
   - 记录 raw 消息延迟；
   - 根据配置检测高延迟/高 RTT，必要时触发重连；
   - 若 `_id` 不为空，执行去重；
   - 通过 `upbitKRWSymbols()` 过滤出 Upbit KRW 相关事件；
   - 调用所有已注册 handler，推送 `Event`.

---

### 五、下游消费建议

1. **注册回调**
   ```go
   treenews.RegisterHandler(func(ctx context.Context, evt treenews.Event) {
       // 例如发送到策略执行、监控系统等
       fmt.Printf("Tree News: id=%s symbols=%v latency=%dms\n", evt.ID, evt.Symbols, evt.LatencyMS)
   })
   ```

2. **利用延迟指标**
   - 可以在 handler 中对 `LatencyMS`/`RTTMS` 进行监控或上报 Prometheus；
   - 对 `Payload` 做自定义解析，提取更多上下文信息。

3. **高延迟处理**
   - 模块已内建高延迟自动重连；若需要额外告警，可在 handler 中自定义推送逻辑（例如调用通知服务）。

---

### 六、测试方式

1. **本地运行**
   ```bash
   go run ./test/tree_news -config=config/config_main.yaml
   ```

2. **实时观察**
   ```bash
   tail -f logs/$(date +%Y-%m-%d)/tree_news_log/tree_news.log
   ```
   日志会显示：
   - `worker=X connected ...` / `login success`
   - `tree news raw msg ... latency=...`
   - `tree news event id=... symbols=[...]`
   - 遇到高延迟、RTT 超限时的 WARN 或重连信息

3. **停用模块**
   - 将 `treeNews.enabled` 改为 `false` 或导出 `TREE_NEWS_ENABLED=0` 即可关闭。

---

### 七、与 Python 版本的差异

- gRPC 发送链路未迁移：若需把 Tree News 推送给策略服务，可在 handler 内自行实现 gRPC/HTTP 客户端；
- 日志格式由 JSON 行日志改为 go logrus（INFO/WARN），路径有所不同；
- 其它特性（延迟监控、去重、滚动重连、意图延时）已全部迁移。

---

若有新的消费场景（例如推送告警、写入数据库等），可直接在 handler 内实现。需要调整队列大小、阈值时修改配置即可生效。*** End Patch
