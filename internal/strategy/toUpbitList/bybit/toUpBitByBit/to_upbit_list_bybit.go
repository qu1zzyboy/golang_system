package toUpBitByBit

import (
	"context"
	"strings"
	"time"
	strategyV1 "upbitBnServer/api/strategy/v1"
	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/infra/errorx/errDefine"
	"upbitBnServer/internal/infra/global/globalCron"
	"upbitBnServer/internal/infra/observe/log/staticLog"
	"upbitBnServer/internal/infra/redisx"
	"upbitBnServer/internal/infra/redisx/redisConfig"
	"upbitBnServer/internal/infra/systemx/instanceEnum"
	"upbitBnServer/internal/quant/exchanges/bybit/autoMarketBybitSub"
	"upbitBnServer/internal/quant/exchanges/bybit/autoMarketChanByBit"
	"upbitBnServer/internal/quant/exchanges/bybit/bybitVar"
	"upbitBnServer/internal/quant/exchanges/bybit/poolMarketBybitSub"
	"upbitBnServer/internal/quant/exchanges/bybit/poolMarketChanByBit"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/quant/market/treeNewsSub"
	"upbitBnServer/internal/resource/resourceEnum"
	"upbitBnServer/internal/strategy/toUpbitList/bybit/toUpbitBybitSymbolArr"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListChan"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitMesh"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitParam"
	"upbitBnServer/pkg/utils/convertx"
	"upbitBnServer/pkg/utils/jsonUtils"
	"upbitBnServer/server/instance/instanceCenter"

	"github.com/tidwall/gjson"
)

const limit = 45

type Req struct {
	PriceRiceTrig float64 // 价格触发阈值,当价格变化超过该值时触发
	OrderRiceTrig float64 // 下单触发阈值,当价格变化超过该值时下单
	Qty           float64 // 开仓金额
	Dec003        float64 // dec0.3
	Dec500        float64 // dec500
}

func (s *Req) TypeName() string {
	return "toUpBitByBit.Req"
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

func (e *Engine) getPreAccountKeyId() uint8 {
	e.thisCalCount++
	if e.thisCalCount == limit {
		e.thisAccountKeyId++
		e.thisCalCount = 0
	}
	return e.thisAccountKeyId
}

func (e *Engine) start(ctx context.Context, req *Req) error {
	toUpBitDataStatic.ExType = exchangeEnum.BYBIT
	toUpBitDataStatic.AcType = exchangeEnum.FUTURE
	toUpbitParam.SetParam(req.Qty, req.Dec003, req.Dec500)
	// --- IGNORE ---
	toUpBitListDataAfter.ClearTrig()

	//获取初始化要订阅的品种
	redisClient, err := redisx.LoadClient(redisConfig.CONFIG_ALL_KEY)
	if err != nil {
		return err
	}
	res := redisClient.HGetAll(ctx, toUpbitMesh.REDIS_KEY_TO_UPBIT_LIST_COIN_BYBIT)
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
		if err = jsonUtils.UnmarshalFromString(v, &mesh); err != nil {
			return err
		}
		if mesh.IsList {
			symbols = append(symbols, mesh.SymbolName)
		}
	}
	if len(symbols) == 0 {
		return toUpBitDataStatic.ExType.GetNotSupportError("no symbols to subscribe")
	}
	toUpBitDataStatic.DyLog.GetLog().Infof("bybit初始化订阅币种:%d,max:%d", len(symbols), len(data))

	// symbols = symbols[:2]

	needLen := len(symbols) + 100
	//chan对象数组
	poolMarketChanByBit.InitChanArr(needLen)
	autoMarketChanByBit.InitChanArr(needLen)
	toUpbitListChan.InitChanArr(needLen)

	//obj对象数组
	toUpbitBybitSymbolArr.InitObjArr(needLen)
	poolMarketBybitSub.InitObjArr(needLen, []resourceEnum.ResourceType{resourceEnum.BOOK_TICK})
	autoMarketBybitSub.InitObjArr(needLen, []resourceEnum.ResourceType{resourceEnum.AGG_TRADE, resourceEnum.MARK_PRICE})

	for index, symbolName := range symbols {
		symbolIndex := bybitVar.GetOrStore(symbolName)
		if err = toUpbitBybitSymbolArr.GetSymbolObj(symbolIndex).Start(e.getPreAccountKeyId(), index, symbolName); err != nil {
			return err
		}
	}

	//初始化各个行情数据引擎
	poolMarketBybitSub.Register(ctx, symbols)
	autoMarketBybitSub.Register(ctx, symbols)

	// 注册treeNews
	if err := treeNewsSub.Get().RegisterReadHandler(ctx, onSymbolPool); err != nil {
		return err
	}
	// 注册交易时间段
	if _, err := globalCron.AddFunc(toUpBitDataStatic.DAY_BEGIN_STR, e.OnDayBegin); err != nil {
		return err
	}
	if _, err := globalCron.AddFunc(toUpBitDataStatic.DAY_END_STR, e.OnDayEnd); err != nil {
		return err
	}
	return nil
}

func onSymbolPool(data []byte) {
	if strings.Contains(strings.ToLower(gjson.GetBytes(data, "source").String()), "upbit") {
		if toUpBitListDataAfter.TrigSymbolIndex != (-1) {
			toUpbitBybitSymbolArr.GetSymbolObj(toUpBitListDataAfter.TrigSymbolIndex).ReceiveTreeNews()
		}
		staticLog.Log.Infof("--->[%d],%s\n", time.Now().UnixMicro(), string(data))
	}
}

func Start(ctx context.Context, meta *strategyV1.ServerReqBase, req *Req) error {
	if instanceCenter.GetManager().IsInstanceExists(instanceEnum.TO_UPBIT_LIST_BYBIT) {
		return errDefine.InstanceExists.WithMetadata(map[string]string{
			defineJson.InstanceId: convertx.ToString(uint32(instanceEnum.TO_UPBIT_LIST_BYBIT)),
			defineJson.Op:         "start",
		})
	}
	s := newEngine()
	if err := s.start(ctx, req); err != nil {
		return err
	}
	if err := toUpbitMesh.GetHandle().Register(ctx, instanceEnum.TO_UPBIT_LIST_BYBIT, map[string]string{}, s); err != nil {
		return err
	}
	if err := instanceCenter.GetManager().Register(ctx, instanceEnum.TO_UPBIT_LIST_BYBIT, map[string]string{}, s); err != nil {
		return err
	}
	return nil
}
