package toUpBitListBn

import (
	"context"

	"github.com/hhh500/quantGoInfra/define/defineJson"
	"github.com/hhh500/quantGoInfra/infra/errorx/errDefine"
	"github.com/hhh500/quantGoInfra/infra/global/globalCron"
	"github.com/hhh500/quantGoInfra/infra/redisx"
	"github.com/hhh500/quantGoInfra/infra/redisx/redisConfig"
	"github.com/hhh500/quantGoInfra/pkg/container/ring/ringBuf"
	"github.com/hhh500/quantGoInfra/pkg/utils/convertx"
	"github.com/hhh500/quantGoInfra/pkg/utils/jsonUtils"
	"github.com/hhh500/quantGoInfra/quant/exchanges/exchangeEnum"
	strategyV1 "github.com/hhh500/upbitBnServer/api/strategy/v1"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/bn/toUpBitListBnAccount"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/bn/toUpBitListBnMarket"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitListBnSymbol"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitListBnSymbolArr"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpbitListChan"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpbitMesh"
	"github.com/hhh500/upbitBnServer/server/instance/instanceCenter"
	"github.com/hhh500/upbitBnServer/server/serverInstanceEnum"
)

const limit = 30

type Req struct {
	PriceRiceTrig float64          // 价格触发阈值,当价格变化超过该值时触发
	OrderRiceTrig float64          // 下单触发阈值,当价格变化超过该值时下单
	Qty           float64          // 开仓金额
	Dec003        float64          // dec0.3
	Dec500        int64            // dec500
	TickCap       ringBuf.Capacity // bookTick环形缓冲区容量
	IsDebug       bool
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

func newEngine() *Engine {
	return &Engine{}
}

func (e *Engine) start(ctx context.Context, req *Req) error {
	toUpBitListDataStatic.SetParam(req.PriceRiceTrig, req.OrderRiceTrig, req.TickCap, req.Dec500, req.IsDebug)
	toUpBitListDataStatic.ExType = exchangeEnum.BINANCE
	toUpBitListDataStatic.AcType = exchangeEnum.FUTURE
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
		if mesh.IsList {
			symbols = append(symbols, mesh.SymbolName)
		}
	}
	if len(symbols) == 0 {
		return toUpBitListDataStatic.ExType.GetNotSupportError("no symbols to subscribe")
	}
	toUpBitListDataStatic.DyLog.GetLog().Infof("bn初始化订阅币种:%d,max:%d", len(symbols), len(data))
	needLen := len(symbols) + 100
	toUpbitListChan.InitUpBit(needLen)
	toUpbitListBnSymbolArr.Init(needLen)

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
	if _, err := globalCron.AddFunc(toUpBitListDataStatic.DAY_BEGIN_STR, e.OnDayBegin); err != nil {
		return err
	}
	if _, err := globalCron.AddFunc(toUpBitListDataStatic.DAY_END_STR, e.OnDayEnd); err != nil {
		return err
	}
	if err := toUpBitListBnAccount.GetBnAccountManager().RefreshSymbolConfig(); err != nil {
		return err
	}
	return nil
}

func Start(ctx context.Context, meta *strategyV1.ServerReqBase, req *Req) error {
	if instanceCenter.GetManager().IsInstanceExists(serverInstanceEnum.TO_UPBIT_LIST_BN) {
		return errDefine.InstanceExists.WithMetadata(map[string]string{
			defineJson.InstanceId: convertx.ToString(uint32(serverInstanceEnum.TO_UPBIT_LIST_BN)),
			defineJson.Op:         "start",
		})
	}
	s := newEngine()
	if err := s.start(ctx, req); err != nil {
		return err
	}
	if err := toUpbitMesh.GetHandle().Register(ctx, serverInstanceEnum.TO_UPBIT_LIST_BN, map[string]string{}, s); err != nil {
		return err
	}
	if err := instanceCenter.GetManager().Register(ctx, serverInstanceEnum.TO_UPBIT_LIST_BN, map[string]string{}, s); err != nil {
		return err
	}
	return nil
}
