package continuousKlineTest

import (
	"context"
	"fmt"
	"time"

	"upbitBnServer/internal/infra/httpx"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/redisx"
	"upbitBnServer/internal/infra/redisx/redisConfig"
	"upbitBnServer/internal/quant/exchanges/binance/bnConst"
	"upbitBnServer/internal/quant/market/kline/klineSubBn"
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
	MODULE_ID             = "continuous_kline_test"
	REDIS_KEY_PREFIX      = "KLINE"                  // Redis Key 前缀（List）
	REDIS_DEDUP_PREFIX    = "KLINE_DEDUP"            // Redis 去重 Key 前缀
	MAX_KLINES_PER_SYMBOL = 600                      // 每个品种最多存储600根K线
	LUA_SCRIPT_KEY        = "continuous_kline_dedup" // Lua 脚本 Key
	DEDUP_KEY_TTL         = 86400                    // 去重 key 过期时间（秒），24小时
)

// Lua 脚本：原子性去重写入（使用 symbol:closeTime 作为去重 key）
// KEYS[1]: listKey (CONTINUOUS_KLINE:{symbol})
// KEYS[2]: dedupKey (CONTINUOUS_KLINE_DEDUP:{symbol}:{closeTime})
// ARGV[1]: klineJson (K线数据 JSON)
// ARGV[2]: maxKlines (最大K线数量)
// ARGV[3]: ttl (去重 key 的过期时间，秒)
// 返回: 1 表示新写入，0 表示重复（已存在）
const luaScriptDedupWrite = `
local listKey = KEYS[1]
local dedupKey = KEYS[2]
local klineJson = ARGV[1]
local maxKlines = tonumber(ARGV[2])
local ttl = tonumber(ARGV[3])

-- 使用 SET key "1" NX 来判断是否已存在
-- 如果 key 不存在，SET 返回 "OK"，如果已存在，返回 nil
local result = redis.call('SET', dedupKey, '1', 'NX', 'EX', ttl)
if result == nil then
    -- key 已存在，返回0表示重复
    return 0
end

-- key 不存在，执行写入操作
redis.call('LPUSH', listKey, klineJson)
redis.call('LTRIM', listKey, 0, maxKlines - 1)

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

	fmt.Printf("[ContinuousKlineTest] 找到 %d 个币安期货 USDT 交易对，开始订阅秒级别连续 K 线...\n", len(symbols))

	// 订阅所有交易对
	err = klineSubBn.GetContinuousKlineManager().RegisterReadHandler(
		ctx,
		symbols,
		s.OnContinuousKLine,
	)
	if err != nil {
		return fmt.Errorf("注册连续 K 线 handler 失败: %w", err)
	}

	fmt.Printf("[ContinuousKlineTest] 订阅成功，等待数据...\n")
	return nil
}

// getAllBinanceFutureUsdtSymbols 从币安 exchangeInfo API 获取所有期货 USDT 交易对
func (s *Service) getAllBinanceFutureUsdtSymbols() []string {
	var symbols []string

	// 调用币安期货 exchangeInfo API
	exchangeInfoUrl := fmt.Sprintf("%s/fapi/v1/exchangeInfo", bnConst.FUTURE_BASE_REST_URL)
	fmt.Printf("[ContinuousKlineTest] 正在从币安获取交易对列表: %s\n", exchangeInfoUrl)

	data, err := httpx.Get(exchangeInfoUrl)
	if err != nil {
		fmt.Printf("[ContinuousKlineTest] 获取币安 exchangeInfo 失败: %v\n", err)
		return symbols
	}

	// 使用 gjson 解析 JSON
	symbolsArray := gjson.GetBytes(data, "symbols")
	if !symbolsArray.Exists() {
		fmt.Printf("[ContinuousKlineTest] exchangeInfo 响应格式错误，未找到 symbols 字段\n")
		return symbols
	}

	// 遍历所有交易对，过滤出 USDT 永续合约
	symbolsArray.ForEach(func(_, symbol gjson.Result) bool {
		// 过滤条件：
		// 1. status == "TRADING" (正在交易)
		// 2. contractType == "PERPETUAL" (永续合约)
		// 3. quoteAsset == "USDT" (USDT 计价)
		status := symbol.Get("status").String()
		contractType := symbol.Get("contractType").String()
		quoteAsset := symbol.Get("quoteAsset").String()
		symbolName := symbol.Get("symbol").String()

		if status == "TRADING" &&
			contractType == "PERPETUAL" &&
			quoteAsset == "USDT" {
			symbols = append(symbols, symbolName)
		}
		return true // 继续遍历
	})

	fmt.Printf("[ContinuousKlineTest] 从币安获取到 %d 个 USDT 永续合约交易对\n", len(symbols))
	return symbols
}

func (s *Service) Stop(ctx context.Context) error {
	fmt.Printf("[ContinuousKlineTest] 停止服务...\n")
	klineSubBn.GetContinuousKlineManager().CloseSub(ctx)
	return nil
}

// OnContinuousKLine 处理连续 K 线数据
// 币安连续 K 线格式：
// {"e":"continuous_kline","E":1607443058651,"ps":"BTCUSDT","ct":"PERPETUAL","k":{...}}
// 或者订阅响应：
// {"result":null,"id":123456}
func (s *Service) OnContinuousKLine(len int, bufPtr *[]byte) {
	data := (*bufPtr)[:len]
	defer byteBufPool.ReleaseBuffer(bufPtr)

	// 检查是否是订阅响应（忽略）
	if gjson.GetBytes(data, "result").Exists() || gjson.GetBytes(data, "id").Exists() {
		// 这是订阅响应，不是 K 线数据，直接返回
		return
	}

	// 检查事件类型是否为 continuous_kline
	eventType := gjson.GetBytes(data, "e").String()
	if eventType != "continuous_kline" {
		// 不是连续 K 线事件，可能是其他消息，打印用于调试
		// fmt.Printf("[ContinuousKlineTest] 未知事件类型: %s, 原始数据: %s\n", eventType, string(data))
		return
	}

	// 解析 K 线数据
	klineData := gjson.GetBytes(data, "k")
	if !klineData.Exists() {
		dynamicLog.Log.GetLog().Warnf("[ContinuousKlineTest] 无法解析 K 线数据，原始数据: %s", string(data))
		return
	}

	// 解析 K 线字段
	// symbol 在外层的 ps 字段，不在 k 对象中
	symbol := gjson.GetBytes(data, "ps").String() // 交易对（Pair）
	openTime := klineData.Get("t").Int()          // 开盘时间
	closeTime := klineData.Get("T").Int()         // 收盘时间
	openPrice := klineData.Get("o").Float()       // 开盘价
	highPrice := klineData.Get("h").Float()       // 最高价
	lowPrice := klineData.Get("l").Float()        // 最低价
	closePrice := klineData.Get("c").Float()      // 收盘价
	volume := klineData.Get("v").Float()          // 成交量
	isClosed := klineData.Get("x").Bool()         // 是否收盘

	// 只处理收盘的K线
	if !isClosed {
		// 调试：打印未收盘的 K 线（每秒可能有很多条，所以只打印前几条）
		// fmt.Printf("[ContinuousKlineTest] 收到未收盘K线 [%s] isClosed=%v\n", symbol, isClosed)
		return
	}

	// 构建K线数据结构（包含品种名称）
	klineRecord := map[string]interface{}{
		"symbol":     symbol,
		"openTime":   openTime,
		"closeTime":  closeTime,
		"openPrice":  openPrice,
		"highPrice":  highPrice,
		"lowPrice":   lowPrice,
		"closePrice": closePrice,
		"volume":     volume,
	}

	// 序列化为 JSON
	klineJson, err := jsonUtils.MarshalStructToString(klineRecord)
	if err != nil {
		dynamicLog.Log.GetLog().Warnf("[ContinuousKlineTest] 序列化K线数据失败 [%s]: %v", symbol, err)
		return
	}

	// 使用 Lua 脚本进行原子性去重写入
	// 去重 key 格式: CONTINUOUS_KLINE_DEDUP:{symbol}:{closeTime}
	listKey := fmt.Sprintf("%s:%s", REDIS_KEY_PREFIX, symbol)
	dedupKey := fmt.Sprintf("%s:%s:%d", REDIS_DEDUP_PREFIX, symbol, closeTime)

	// 执行 Lua 脚本
	// KEYS[1]: listKey, KEYS[2]: dedupKey
	// ARGV[1]: klineJson, ARGV[2]: maxKlines, ARGV[3]: ttl
	result, err := s.luaScript.Eval(s.ctx, []string{listKey, dedupKey}, klineJson, MAX_KLINES_PER_SYMBOL, DEDUP_KEY_TTL)
	if err != nil {
		dynamicLog.Log.GetLog().Warnf("[ContinuousKlineTest] Lua脚本执行失败 [%s]: %v", symbol, err)
		return
	}

	// result 为 1 表示新写入，0 表示重复（已存在）
	if resultInt, ok := result.(int64); ok {
		if resultInt == 1 {
			// 新写入，使用 debug 级别日志
			dynamicLog.Log.GetLog().Debugf("[%s] %s | O:%.2f H:%.2f L:%.2f C:%.2f | Vol:%.2f | 已存储到Redis",
				time.Now().Format("15:04:05"),
				symbol,
				openPrice,
				highPrice,
				lowPrice,
				closePrice,
				volume,
			)
		}
		// resultInt == 0 表示重复，不打印日志（避免日志过多）
	} else {
		dynamicLog.Log.GetLog().Warnf("[ContinuousKlineTest] Lua脚本返回结果类型错误 [%s]: %v", symbol, result)
	}
}
