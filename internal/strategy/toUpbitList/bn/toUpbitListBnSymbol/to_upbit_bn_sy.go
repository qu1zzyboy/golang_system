package toUpbitListBnSymbol

import (
	"context"
	"sync/atomic"
	"time"
	"upbitBnServer/internal/quant/exchanges/binance/marketSub/bnPoolMarketChan"
	"upbitBnServer/internal/quant/exchanges/binance/order/bnOrderTemplate"

	"upbitBnServer/internal/conf"
	"upbitBnServer/internal/infra/latency"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/quant/market/symbolInfo"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolDynamic"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolLimit"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolStatic"
	"upbitBnServer/internal/resource/resourceEnum"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitPointPreBn"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListChan"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListPos"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitParam"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitPoint"
	"upbitBnServer/pkg/container/map/myMap"
	"upbitBnServer/pkg/container/pool/byteBufPool"
	"upbitBnServer/pkg/utils/idGen"

	"github.com/shopspring/decimal"
)

const (
	total = "BINANCE_TOTAL"
)

type cache_line_1 struct {
	symbolName      string
	priceMaxBuy     float64                  // 价格上限
	thisMarkPrice   float64                  // 上次标记价格
	upLimitPercent  float64                  // 涨停百分比,115就是1.15
	minPriceAfterMp float64                  // 标记价格之后的最小ask
	markPriceTs     int64                    // 接受到标记价格的时间
	cmcId           uint32                   //
	symbolLen       uint16                   //
	upT             toUpbitPoint.UpLimitType // 价格限制类型
}
type cache_line_2 struct {
	btLatencyTotal latency.Latency // 延迟统计
	lastRiseValue  float64         // 上次涨幅
	thisMinTs      int64           // 当前分钟的时间戳
	last2MinClose  float64         // 最近两分钟的收盘价
	last1MinClose  float64         // 最近一分钟的收盘价
	thisMinClose   float64         // 当前分钟的收盘价
}

type cache_line_3 struct {
	bidPrice        atomic.Value           // 买一价,平仓和计算仓位价值用到
	takeProfitPrice float64                // 止盈价格
	symbolIndex     systemx.SymbolIndex16I // 交易对下标
	pScale          systemx.PScale         // 价格小数位
	qScale          systemx.QScale         // 数量小数位
	hasAllFilled    atomic.Bool            // 是否已经完全成交
	isStopLossAble  atomic.Bool            // 能否开始移动止损
	hasReceiveStop  bool                   // 是否已经收到过停止信号
	hasTreeNews     bool                   // 是否已经接受到treeNews
}

type cache_line_4 struct {
	orderNum     decimal.Decimal // 预挂单订单数量
	smallPercent decimal.Decimal // 小订单比例
	hasInit      bool            // 是否已经预挂单初始化
}
type cache_line_5 struct {
	chanPoolMarket chan systemx.Job // 盘口数据chan
}

type cache_line_6 struct {
	trigPriceMax myMap.MySyncMap[int64, float64] // 已触发品种的买入上限,秒级别时间戳和价格
}

type Single struct {
	cache_line_1
	cache_line_2
	cache_line_3
	cache_line_4
	cache_line_5
	cache_line_6
	chanTrigOrder      chan toUpbitListChan.TrigOrderInfo      // 简易成交数据chan
	chanMonitor        chan toUpbitListChan.MonitorResp        // 订单监测chan
	chanOutSideSig     chan toUpbitListChan.Special            // 外部信号chan
	chanSuOrder        chan toUpBitListDataAfter.OnSuccessEvt  // 成功订单chan
	secondArr          [11]*secondPerInfo                      // 每秒信息
	clientOrderIds     myMap.MySyncMap[systemx.WsId16B, uint8] // 挂单成功的clientOrderId
	posTotalNeed       float64                                 // 需要开仓的数量
	maxNotional        float64                                 // 单品种最大开仓上限
	ctxStop            context.Context                         // 同步关闭ctx
	cancel             context.CancelFunc                      // 关闭函数
	pos                *toUpbitListPos.PosCalSafe              // 持仓计算对象
	pre                *toUpbitPointPreBn.PointPre             // 预挂单对象
	can                *bnOrderTemplate.CancelTemplate         // 撤单json模板
	twapSec            float64                                 // twap下单间隔秒数
	closeDuration      time.Duration                           // 平仓持续时间
	thisOrderAccountId atomic.Int32                            // 当前订单使用的资金账户ID
	toAccountId        atomic.Int32                            // 准备接收资金的账户id
}

func (s *Single) Start(accountKeyId uint8, index int, symbolName string) error {
	s.symbolName = symbolName
	s.symbolIndex = systemx.SymbolIndex16I(index)
	s.symbolLen = uint16(len(symbolName))

	symbolId, ok := symbolStatic.GetSymbol().GetSymbol(symbolName)
	if !ok {
		toUpBitDataStatic.DyLog.GetLog().Errorf("symbolId not found: %s", symbolName)
		return toUpBitDataStatic.ExType.GetNotSupportError("symbolId not found")
	}
	symbolKeyId := symbolInfo.MakeSymbolKey3(toUpBitDataStatic.ExType, toUpBitDataStatic.AcType, symbolId)
	// 初始化品种静态数据
	stMeta, err := symbolStatic.GetTrade().Get(symbolKeyId)
	if err != nil {
		toUpBitDataStatic.DyLog.GetLog().Errorf("symbolKeyId %d not found", symbolKeyId)
		return err
	}
	s.cmcId = stMeta.TradeId
	s.pre = toUpbitPointPreBn.NewPre(accountKeyId, s.symbolLen, s.symbolIndex, stMeta.SymbolKeyId)
	s.can = bnOrderTemplate.NewCancelTemplate()
	s.can.Start(symbolName)

	// 初始化品种动态数据
	dyMeta, err := symbolDynamic.GetManager().Get(symbolKeyId)
	if err != nil {
		toUpBitDataStatic.DyLog.GetLog().Errorf("symbolKeyId %d not found", symbolKeyId)
		return err
	}
	s.pScale = systemx.PScale(dyMeta.PScale)
	s.qScale = systemx.QScale(dyMeta.QScale)

	limit, err := symbolLimit.GetManager().Get(symbolKeyId)
	if err != nil {
		toUpBitDataStatic.DyLog.GetLog().Errorf("symbolKeyId %d not found", symbolKeyId)
		return err
	}

	// 1.15-->115
	s.upLimitPercent = limit.UpLimitPercent.InexactFloat64()

	s.chanPoolMarket = make(chan systemx.Job, 100)
	bnPoolMarketChan.Register(s.symbolIndex, s.chanPoolMarket)

	s.chanTrigOrder = make(chan toUpbitListChan.TrigOrderInfo, 10)
	s.chanMonitor = make(chan toUpbitListChan.MonitorResp, 10)
	s.chanOutSideSig = make(chan toUpbitListChan.Special, 10)
	s.chanSuOrder = make(chan toUpBitListDataAfter.OnSuccessEvt, 10)

	toUpbitListChan.RegisterSpecial(s.symbolIndex, s.chanOutSideSig)
	toUpbitListChan.RegisterBnPrivate(s.symbolIndex, s.chanTrigOrder, s.chanMonitor, s.chanSuOrder)
	latencyPrefix := idGen.BuildName2("GO", conf.ServerName)
	s.btLatencyTotal = latency.NewHttpMonitor(idGen.BuildName2(latencyPrefix, total), latency.PROCESS_TOTAL, resourceEnum.BOOK_TICK)
	for i := range toUpbitParam.MaxAccount {
		temp := &secondPerInfo{}
		temp.clear()
		s.secondArr[i] = temp
	}
	s.trigPriceMax = myMap.NewMySyncMap[int64, float64]()
	safex.SafeGo(symbolName+"单品种协程", s.onLoop)
	return nil
}

func (s *Single) onLoop() {
	for {
		select {
		case job := <-s.chanPoolMarket:
			s.handleMarketJob(job)
		case job := <-s.chanTrigOrder:
			s.onTradeLitePre(job)
		case job := <-s.chanMonitor:
			s.priceMaxBuy = job.P
			s.trigPriceMax.Store(time.Now().Unix(), job.P)
			toUpBitDataStatic.DyLog.GetLog().Infof("最新探测价格[%.8f]", job.P)
		case sig := <-s.chanOutSideSig:
			{
				switch sig.SigType {
				case toUpbitListChan.ReceiveTreeNews:
					{
						s.ReceiveTreeNews()
					}
				case toUpbitListChan.ReceiveNoTreeNews:
					{
						s.ReceiveNoTreeNews()
					}
				case toUpbitListChan.ReceiveTreeNewsAndOpen:
					{
						s.intoExecuteByMsg()
					}
				case toUpbitListChan.CancelOrderReturn:
					{
						s.onCanceledOrder(sig.AccountKeyId)
					}
				case toUpbitListChan.FailureOrder:
					{
						s.onFailureOrder(sig.AccountKeyId, sig.ErrCode)
					}
				case toUpbitListChan.QUERY_ACCOUNT_RETURN:
					{
						s.onMaxWithdrawAmount(sig.AccountKeyId, sig.Amount)
					}
				}
			}
		case evt := <-s.chanSuOrder:
			s.onSuccessOrder(evt)
		}
	}
}

func (s *Single) handleMarketJob(job systemx.Job) {
	defer byteBufPool.ReleaseBuffer(job.Buf)
	b := (*job.Buf)[:job.Len]
	switch {
	case b[6] == 'm' && b[7] == 'a' && b[8] == 'r' && b[9] == 'k':
		s.onMarkPrice(job.Len, b)
	case b[6] == 'a' && b[7] == 'g' && b[8] == 'g' && b[9] == 'T':
		// s.onAggTrade(job.Len, b)
	case b[6] == 'b' && b[7] == 'o' && b[8] == 'o' && b[9] == 'k':
		s.onBookTick(job.Len, b)
	default:
		dynamicLog.Error.GetLog().Errorf("err json: %s", string(b))
	}
}
