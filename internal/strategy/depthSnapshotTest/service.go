package depthSnapshotTest

import (
	"context"
	"fmt"
	"time"

	"upbitBnServer/internal/infra/httpx"
	"upbitBnServer/internal/infra/redisx"
	"upbitBnServer/internal/infra/redisx/redisConfig"
	"upbitBnServer/internal/quant/exchanges/binance/bnConst"
	"upbitBnServer/internal/quant/market/depth/depthModel"
	"upbitBnServer/internal/quant/market/depth/depthSubBn"
	"upbitBnServer/pkg/container/pool/byteBufPool"
	"upbitBnServer/pkg/singleton"
	"upbitBnServer/pkg/utils/jsonUtils"

	"github.com/go-redis/redis/v8"
	"github.com/tidwall/gjson"
)

var (
	serviceSingleton = singleton.NewSingleton(func() *Service { return newService() })
)

const (
	MODULE_ID                = "depth_snapshot_test"
	REDIS_KEY_PREFIX         = "DEPTH"                // Redis Key 前缀（List）
	REDIS_DEDUP_PREFIX       = "DEPTH_DEDUP"          // Redis 去重 Key 前缀
	MAX_SNAPSHOTS_PER_SYMBOL = 1200                   // 每个品种最多存储600个快照
	LUA_SCRIPT_KEY           = "depth_snapshot_dedup" // Lua 脚本 Key
	DEDUP_KEY_TTL            = 86400                  // 去重 key 过期时间（秒），24小时
	DEPTH_LEVELS             = 20                     // 订单簿深度档位
	UPDATE_SPEED_MS          = 500                    // 更新速度（毫秒）
)

// Lua 脚本：原子性去重写入（使用 symbol:eventTime 作为去重 key）
// KEYS[1]: listKey (DEPTH:{symbol})
// KEYS[2]: dedupKey (DEPTH_DEDUP:{symbol}:{eventTime})
// ARGV[1]: snapshotJson (订单簿快照 JSON)
// ARGV[2]: maxSnapshots (最大快照数量)
// ARGV[3]: ttl (去重 key 的过期时间，秒)
// 返回: 1 表示新写入，0 表示重复（已存在）
const luaScriptDedupWrite = `
local listKey = KEYS[1]
local dedupKey = KEYS[2]
local snapshotJson = ARGV[1]
local maxSnapshots = tonumber(ARGV[2])
local ttl = tonumber(ARGV[3])

-- 使用 SET key "1" NX 来判断是否已存在
local result = redis.call('SET', dedupKey, '1', 'NX', 'EX', ttl)
if result == nil then
    -- key 已存在，返回0表示重复
    return 0
end

-- key 不存在，执行写入操作
redis.call('LPUSH', listKey, snapshotJson)
redis.call('LTRIM', listKey, 0, maxSnapshots - 1)

-- 返回1表示新写入
return 1
`

func GetService() *Service {
	return serviceSingleton.Get()
}

type Service struct {
	ctx         context.Context
	redisClient *redis.Client
	luaScript   *redisx.LuaScript
}

func newService() *Service {
	return &Service{}
}

func (s *Service) Start(ctx context.Context) error {
	s.ctx = ctx

	// 加载 Redis 客户端
	var err error
	s.redisClient, err = redisx.LoadClient(redisConfig.CONFIG_ALL_KEY)
	if err != nil {
		return fmt.Errorf("加载 Redis 客户端失败: %w", err)
	}

	// 注册 Lua 脚本用于去重写入
	err = redisx.RegisterLuaScript(ctx, s.redisClient, LUA_SCRIPT_KEY, luaScriptDedupWrite)
	if err != nil {
		return fmt.Errorf("注册 Lua 脚本失败: %w", err)
	}

	// 加载 Lua 脚本
	s.luaScript, err = redisx.LoadLuaScript(LUA_SCRIPT_KEY)
	if err != nil {
		return fmt.Errorf("加载 Lua 脚本失败: %w", err)
	}

	// 获取所有币安期货 USDT 交易对
	symbols := s.getAllBinanceFutureUsdtSymbols()
	if len(symbols) == 0 {
		return fmt.Errorf("未找到任何币安期货 USDT 交易对")
	}

	fmt.Printf("[DepthSnapshotTest] 找到 %d 个币安期货 USDT 交易对，开始订阅500ms订单簿快照...\n", len(symbols))

	// 订阅所有交易对
	err = depthSubBn.GetManager().RegisterReadHandler(
		ctx,
		symbols,
		s.OnDepthSnapshot,
	)
	if err != nil {
		return fmt.Errorf("注册订单簿快照 handler 失败: %w", err)
	}

	fmt.Printf("[DepthSnapshotTest] 订阅成功，等待数据...\n")
	return nil
}

// getAllBinanceFutureUsdtSymbols 从币安 exchangeInfo API 获取所有期货 USDT 交易对
func (s *Service) getAllBinanceFutureUsdtSymbols() []string {
	var symbols []string

	// 调用币安期货 exchangeInfo API
	exchangeInfoUrl := fmt.Sprintf("%s/fapi/v1/exchangeInfo", bnConst.FUTURE_BASE_REST_URL)
	fmt.Printf("[DepthSnapshotTest] 正在从币安获取交易对列表: %s\n", exchangeInfoUrl)

	data, err := httpx.Get(exchangeInfoUrl)
	if err != nil {
		fmt.Printf("[DepthSnapshotTest] 获取币安 exchangeInfo 失败: %v\n", err)
		return symbols
	}

	// 使用 gjson 解析 JSON
	symbolsArray := gjson.GetBytes(data, "symbols")
	if !symbolsArray.Exists() {
		fmt.Printf("[DepthSnapshotTest] exchangeInfo 响应格式错误，未找到 symbols 字段\n")
		return symbols
	}

	// 遍历所有交易对，过滤出 USDT 永续合约
	symbolsArray.ForEach(func(_, symbol gjson.Result) bool {
		status := symbol.Get("status").String()
		contractType := symbol.Get("contractType").String()
		quoteAsset := symbol.Get("quoteAsset").String()
		symbolName := symbol.Get("symbol").String()

		if status == "TRADING" &&
			contractType == "PERPETUAL" &&
			quoteAsset == "USDT" {
			symbols = append(symbols, symbolName)
		}
		return true
	})

	fmt.Printf("[DepthSnapshotTest] 从币安获取到 %d 个 USDT 永续合约交易对\n", len(symbols))
	return symbols
}

func (s *Service) Stop(ctx context.Context) error {
	fmt.Printf("[DepthSnapshotTest] 停止服务...\n")
	depthSubBn.GetManager().CloseSub(ctx)
	return nil
}

// OnDepthSnapshot 处理订单簿深度快照数据
// 币安订单簿深度格式：
// {"e":"depthUpdate","E":1607443058651,"s":"BTCUSDT","U":123456,"u":123460,"b":[...],"a":[...]}
func (s *Service) OnDepthSnapshot(dataLen int, bufPtr *[]byte) {
	data := (*bufPtr)[:dataLen]
	defer byteBufPool.ReleaseBuffer(bufPtr)

	// 检查是否是订阅响应（忽略）
	if gjson.GetBytes(data, "result").Exists() || gjson.GetBytes(data, "id").Exists() {
		return
	}

	// 检查事件类型是否为 depthUpdate
	eventType := gjson.GetBytes(data, "e").String()
	if eventType != "depthUpdate" {
		// 调试：打印所有非订阅响应的消息（前10条）
		if len(string(data)) < 500 { // 只打印短消息，避免日志过多
			fmt.Printf("[DepthSnapshotTest] 未知事件类型: %s, 原始数据: %s\n", eventType, string(data))
		}
		return
	}

	// 调试：打印收到的 depthUpdate 事件（前几条，只打印前200字符）
	dataStr := string(data)
	if len(dataStr) > 200 {
		fmt.Printf("[DepthSnapshotTest] 收到 depthUpdate 事件: %s...\n", dataStr[:200])
	} else {
		fmt.Printf("[DepthSnapshotTest] 收到 depthUpdate 事件: %s\n", dataStr)
	}

	// 解析订单簿数据到 model
	depthUpdate := depthModel.DepthUpdate{
		Symbol:            gjson.GetBytes(data, "s").String(), // 交易对
		EventTime:         gjson.GetBytes(data, "E").Int(),    // 事件时间
		TransactionTime:   gjson.GetBytes(data, "T").Int(),    // 交易时间
		FirstUpdateId:     gjson.GetBytes(data, "U").Int(),    // 首次更新ID
		FinalUpdateId:     gjson.GetBytes(data, "u").Int(),    // 最终更新ID
		PrevFinalUpdateId: gjson.GetBytes(data, "pu").Int(),   // 上次流的最终更新ID
	}

	// 解析 bids 和 asks（二维数组）
	bidsJson := gjson.GetBytes(data, "b")
	asksJson := gjson.GetBytes(data, "a")

	bidsArray := bidsJson.Array()
	asksArray := asksJson.Array()

	depthUpdate.Bids = make([][]string, 0, len(bidsArray))
	depthUpdate.Asks = make([][]string, 0, len(asksArray))

	for _, bid := range bidsArray {
		bidArray := bid.Array()
		if len(bidArray) >= 2 {
			depthUpdate.Bids = append(depthUpdate.Bids, []string{
				bidArray[0].String(), // 价格
				bidArray[1].String(), // 数量
			})
		}
	}

	for _, ask := range asksArray {
		askArray := ask.Array()
		if len(askArray) >= 2 {
			depthUpdate.Asks = append(depthUpdate.Asks, []string{
				askArray[0].String(), // 价格
				askArray[1].String(), // 数量
			})
		}
	}

	// 构建订单簿快照数据结构（用于 Redis 存储）
	snapshotRecord := map[string]interface{}{
		"symbol":            depthUpdate.Symbol,
		"eventTime":         depthUpdate.EventTime,
		"transactionTime":   depthUpdate.TransactionTime,
		"firstUpdateId":     depthUpdate.FirstUpdateId,
		"finalUpdateId":     depthUpdate.FinalUpdateId,
		"prevFinalUpdateId": depthUpdate.PrevFinalUpdateId,
		"bids":              depthUpdate.Bids,
		"asks":              depthUpdate.Asks,
	}

	// 序列化为 JSON
	snapshotJson, err := jsonUtils.MarshalStructToString(snapshotRecord)
	if err != nil {
		fmt.Printf("[DepthSnapshotTest] 序列化订单簿快照失败: %v\n", err)
		return
	}

	// 使用 Lua 脚本进行原子性去重写入
	listKey := fmt.Sprintf("%s:%s", REDIS_KEY_PREFIX, depthUpdate.Symbol)
	dedupKey := fmt.Sprintf("%s:%s:%d", REDIS_DEDUP_PREFIX, depthUpdate.Symbol, depthUpdate.EventTime)

	// 执行 Lua 脚本
	result, err := s.luaScript.Eval(s.ctx, []string{listKey, dedupKey}, snapshotJson, MAX_SNAPSHOTS_PER_SYMBOL, DEDUP_KEY_TTL)
	if err != nil {
		fmt.Printf("[DepthSnapshotTest] Lua脚本执行失败 [%s]: %v\n", depthUpdate.Symbol, err)
		return
	}

	// result 为 1 表示新写入，0 表示重复（已存在）
	if resultInt, ok := result.(int64); ok {
		if resultInt == 1 {
			// 新写入，打印日志
			fmt.Printf("[%s] %s | EventTime:%d | FirstU:%d FinalU:%d PrevU:%d | Bids:%d Asks:%d | 已存储到Redis\n",
				time.Now().Format("15:04:05"),
				depthUpdate.Symbol,
				depthUpdate.EventTime,
				depthUpdate.FirstUpdateId,
				depthUpdate.FinalUpdateId,
				depthUpdate.PrevFinalUpdateId,
				len(depthUpdate.Bids),
				len(depthUpdate.Asks),
			)
		}
		// resultInt == 0 表示重复，不打印日志（避免日志过多）
	} else {
		fmt.Printf("[DepthSnapshotTest] Lua脚本返回结果类型错误 [%s]: %v\n", depthUpdate.Symbol, result)
	}
}
