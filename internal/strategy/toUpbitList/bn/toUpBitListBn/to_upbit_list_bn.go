package toUpBitListBn

import (
	"context"
	"upbitBnServer/internal/quant/exchanges/binance/marketSub/bnPoolMarketChan"

	strategyV1 "upbitBnServer/api/strategy/v1"
	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/infra/errorx/errDefine"
	"upbitBnServer/internal/infra/global/globalCron"
	"upbitBnServer/internal/infra/redisx"
	"upbitBnServer/internal/infra/redisx/redisConfig"
	"upbitBnServer/internal/infra/systemx/instanceEnum"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpBitListBnAccount"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpBitListBnMarket"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitListBnSymbol"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitListBnSymbolArr"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListChan"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitMesh"
	"upbitBnServer/pkg/container/ring/ringBuf"
	"upbitBnServer/pkg/utils/convertx"
	"upbitBnServer/pkg/utils/jsonUtils"
	"upbitBnServer/server/instance/instanceCenter"
)

const limit = 30

type Req struct {
	PriceRiceTrig float64          // 价格触发阈值,当价格变化超过该值时触发
	Qty           float64          // 开仓金额
	Dec003        float64          // dec0.3
	Dec500        int64            // dec500
	TickCap       ringBuf.Capacity // bookTick环形缓冲区容量
}

func (s *Req) TypeName() string {
	return "toUpBitListBn.Req"
}

func (s *Req) Check() error {
	return nil
}

type Engine struct {
	thisCalCount     int32 // 当前计算次数
	thisAccountKeyId uint8 // 账户KeyId
}

func (e *Engine) getPreAccountKeyId() uint8 {
	e.thisCalCount++
	if e.thisCalCount == limit {
		e.thisAccountKeyId++
		e.thisCalCount = 0
	}
	return e.thisAccountKeyId
}

func newEngine() *Engine {
	return &Engine{}
}

func (e *Engine) start(ctx context.Context, req *Req) error {
	toUpBitDataStatic.SetParam(req.PriceRiceTrig, req.TickCap, req.Dec500)
	toUpBitDataStatic.ExType = exchangeEnum.BINANCE
	toUpBitDataStatic.AcType = exchangeEnum.FUTURE
	toUpbitListBnSymbol.SetParam(req.Qty, req.Dec003)
	// --- IGNORE ---
	toUpBitListDataAfter.ClearTrig()
	//获取初始化要订阅的品种
	redisClient, err := redisx.LoadClient(redisConfig.CONFIG_ALL_KEY)
	if err != nil {
		return err
	}
	res := redisClient.HGetAll(ctx, toUpbitMesh.REDIS_KEY_TO_UPBIT_LIST_COIN_BN)
	if res.Err() != nil {
		return res.Err()
	}
	data, err := res.Result()
	if err != nil {
		return err
	}
	// 要订阅的品种
	var symbols []string
	for _, v := range data {
		var mesh toUpbitMesh.Save
		if err := jsonUtils.UnmarshalFromString(v, &mesh); err != nil {
			return err
		}
		if mesh.SymbolName == "AI16ZUSDT" || mesh.SymbolName == "SLERFUSDT" {
			continue
		}
		if mesh.IsList {
			symbols = append(symbols, mesh.SymbolName)
		}
	}
	if len(symbols) == 0 {
		return toUpBitDataStatic.ExType.GetNotSupportError("no symbols to subscribe")
	}
	toUpBitDataStatic.DyLog.GetLog().Infof("bn初始化订阅币种:%d,max:%d", len(symbols), len(data))
	needLen := len(symbols) + 100

	//chan对象数组
	bnPoolMarketChan.InitChanArr(needLen)
	toUpbitListChan.InitChanArr(needLen)

	for index, symbolName := range symbols {
		// 提前指定每个品种的挂单账户
		e.thisCalCount++
		if e.thisCalCount == limit {
			e.thisAccountKeyId++
			e.thisCalCount = 0
		}

		if err := toUpbitListBnSymbolArr.GetSymbolObj(index).Start(e.thisAccountKeyId, index, symbolName); err != nil {
			return err
		}
	}
	// 注册行情资源
	if err := toUpBitListBnMarket.GetMarket().RegisterBefore(ctx, symbols); err != nil {
		return err
	}
	// 注册交易时间段
	if _, err := globalCron.AddFunc(toUpBitDataStatic.DAY_BEGIN_STR, e.OnDayBegin); err != nil {
		return err
	}
	if _, err := globalCron.AddFunc(toUpBitDataStatic.DAY_END_STR, e.OnDayEnd); err != nil {
		return err
	}
	if err := toUpBitListBnAccount.GetBnAccountManager().RefreshSymbolConfig(); err != nil {
		return err
	}
	return nil
}

func Start(ctx context.Context, meta *strategyV1.ServerReqBase, req *Req) error {
	if instanceCenter.GetManager().IsInstanceExists(instanceEnum.TO_UPBIT_LIST_BN) {
		return errDefine.InstanceExists.WithMetadata(map[string]string{
			defineJson.InstanceId: convertx.ToString(uint32(instanceEnum.TO_UPBIT_LIST_BN)),
			defineJson.Op:         "start",
		})
	}
	s := newEngine()
	if err := s.start(ctx, req); err != nil {
		return err
	}
	if err := toUpbitMesh.GetHandle().Register(ctx, instanceEnum.TO_UPBIT_LIST_BN, map[string]string{}, s); err != nil {
		return err
	}
	if err := instanceCenter.GetManager().Register(ctx, instanceEnum.TO_UPBIT_LIST_BN, map[string]string{}, s); err != nil {
		return err
	}
	return nil
}
