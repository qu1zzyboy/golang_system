package toUpbitBybitSymbol

import (
	"context"
	"sync/atomic"
	"time"
	"upbitBnServer/internal/conf"
	"upbitBnServer/internal/infra/latency"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/quant/exchanges/bybit/marketSub/bybitAutoMarketChan"
	"upbitBnServer/internal/quant/exchanges/bybit/marketSub/bybitPoolMarketChan"
	"upbitBnServer/internal/quant/market/symbolInfo"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolDynamic"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolLimit"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolStatic"
	"upbitBnServer/internal/resource/resourceEnum"
	"upbitBnServer/internal/strategy/toUpbitList/bybit/toUpbitPointPreByBit"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListChan"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListPos"
	"upbitBnServer/pkg/container/map/myMap"
	"upbitBnServer/pkg/container/pool/byteBufPool"
	"upbitBnServer/pkg/utils/idGen"
)

const (
	total = "BYBIT_TOTAL"
)

type cache_line_1 struct {
	symbolName     string
	priceMaxBuy    float64 // 价格上限(放大了10的10次方倍)
	upLimitPercent float64 // 涨停百分比,115就是1.15
}

type cache_line_2 struct {
	btLatencyTotal  latency.Latency // 延迟统计
	committedTs     int64           // 【已提交】最后一次成功写入的 ts
	thisMinTs       int64           // 当前分钟的时间戳
	last2MinClose_8 uint64          // 最近两分钟的收盘价
	last1MinClose_8 uint64          // 最近一分钟的收盘价
	thisMinClose_8  uint64          // 当前分钟的收盘价
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

type cache_line_5 struct {
	agLatencyTotal latency.Latency  // 延迟统计
	chanMarketPool chan systemx.Job // 盘口数据chan
}

type cache_line_6 struct {
	trigPriceMax myMap.MySyncMap[int64, float64] // 已触发品种的买入上限,秒级别时间戳和价格
}

type Single struct {
	cache_line_1
	cache_line_2
	cache_line_3
	cache_line_5
	cache_line_6
	symbolLen      uint16
	chanMarketAuto chan []byte                             // 行情数据自动chan
	chanTrigOrder  chan toUpbitListChan.TrigOrderInfo      // 简易成交数据chan
	chanOutSideSig chan toUpbitListChan.Special            // 外部信号chan
	chanSuOrder    chan toUpBitListDataAfter.OnSuccessEvt  // 成功订单chan
	clientOrderIds myMap.MySyncMap[systemx.WsId16B, uint8] // 挂单成功的clientOrderId
	posTotalNeed   float64                                 // 需要开仓的数量
	maxNotional    float64                                 // 单品种最大开仓上限
	ctxStop        context.Context                         // 同步关闭ctx
	cancel         context.CancelFunc                      // 关闭函数
	posLong        *toUpbitListPos.PosCalSafe              // 持仓计算对象
	pre            *toUpbitPointPreByBit.PointPre          // 预挂单对象
	twapSec        float64                                 // twap下单间隔秒数
	closeDuration  time.Duration                           // 平仓持续时间
}

func (s *Single) Clear() {
}

func (s *Single) ClearBegin() {
	if s.pre != nil {
		s.pre.FreshClientOrderId()
	}
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
		toUpBitDataStatic.DyLog.GetLog().Errorf("symbolKeyId %d not found for %s", symbolKeyId, symbolName)
		return err
	}
	s.pre = toUpbitPointPreByBit.NewPre(accountKeyId, s.symbolLen, s.symbolIndex, stMeta.SymbolKeyId)
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
	// 0.15-->115
	s.upLimitPercent = 1 + limit.UpLimitPercent.InexactFloat64()
	s.chanMarketPool = make(chan systemx.Job, 100)
	bybitPoolMarketChan.Register(s.symbolIndex, s.chanMarketPool)

	s.chanMarketAuto = make(chan []byte, 100)
	bybitAutoMarketChan.Register(s.symbolIndex, s.chanMarketAuto)

	s.chanTrigOrder = make(chan toUpbitListChan.TrigOrderInfo, 10)
	s.chanOutSideSig = make(chan toUpbitListChan.Special, 10)
	s.chanSuOrder = make(chan toUpBitListDataAfter.OnSuccessEvt, 10)
	toUpbitListChan.RegisterSpecial(s.symbolIndex, s.chanOutSideSig)
	toUpbitListChan.RegisterByBitPrivate(s.symbolIndex, s.chanTrigOrder, s.chanSuOrder)
	latencyPrefix := idGen.BuildName2("GO", conf.ServerName)
	s.agLatencyTotal = latency.NewHttpMonitor(idGen.BuildName2(latencyPrefix, total), latency.PROCESS_TOTAL, resourceEnum.AGG_TRADE)
	s.btLatencyTotal = latency.NewHttpMonitor(idGen.BuildName2(latencyPrefix, total), latency.PROCESS_TOTAL, resourceEnum.BOOK_TICK)
	s.trigPriceMax = myMap.NewMySyncMap[int64, float64]()
	safex.SafeGo(symbolName+"单品种协程", s.onLoop)
	return nil
}

func (s *Single) onLoop() {
	for {
		select {
		case data := <-s.chanMarketAuto:
			switch {
			case data[10] == 'p' && data[11] == 'u':
				s.onAggTrade(data)
			case data[10] == 't' && data[11] == 'i':
				s.onMarkPrice(data)
			default:
				toUpBitDataStatic.DyLog.GetLog().Errorf("未知json:%s", string(data))
			}
		case job := <-s.chanMarketPool:
			s.handleBookTick(job)
		case job := <-s.chanTrigOrder:
			s.onTradeLite(job)
		case evt := <-s.chanSuOrder:
			s.onSuccessOrder(evt)
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
				case toUpbitListChan.FailureOrder:
					{
						s.onFailureOrder(sig.AccountKeyId, sig.ErrCode)
					}
				}
			}
		case evt := <-s.chanSuOrder:
			s.onSuccessOrder(evt)
		}
	}
}

func (s *Single) handleBookTick(job systemx.Job) {
	defer byteBufPool.ReleaseBuffer(job.Buf)
	b := (*job.Buf)[:job.Len]
	s.onBookTick(job.Len, b)
}
