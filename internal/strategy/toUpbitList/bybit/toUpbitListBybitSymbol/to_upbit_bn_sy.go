package toUpbitListBybitSymbol

import (
	"context"
	"sync/atomic"
	"time"
	"upbitBnServer/internal/infra/systemx"

	"upbitBnServer/internal/conf"
	"upbitBnServer/internal/infra/latency"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"upbitBnServer/internal/quant/market/symbolInfo"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolDynamic"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolStatic"
	"upbitBnServer/internal/resource/resourceEnum"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListChan"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListPos"
	"upbitBnServer/pkg/container/map/myMap"
	"upbitBnServer/pkg/utils/convertx"
	"upbitBnServer/pkg/utils/idGen"

	"github.com/shopspring/decimal"
)

const (
	total                    = "BINANCE_TOTAL"
	jsonEvent                = "E"
	jsonT                    = "T"
	upPercentZoomScale int32 = 2 // 价格放大的倍数,10的ZoomScale次方倍
	ws_req_from              = orderBelongEnum.TO_UPBIT_LIST_PRE
)

type cache_line_1 struct {
	mpLatencyTotal   latency.Latency // 延迟统计
	priceMaxBuy_10   uint64          // 价格上限(放大了10的10次方倍)
	lastMarkPrice_8  uint64          // 上次标记价格
	markPrice_8      uint64          // 最新标记价格(放大了10的8次方倍)
	upLimitPercent_2 uint64          // 涨停百分比,115就是1.15
	minPriceAfterMp  uint64          // 标记价格之后的最小ask
	markPriceTs      int64           // 接受到标记价格的时间
}
type cache_line_2 struct {
	btLatencyTotal  latency.Latency // 延迟统计
	committedTs     int64           // 【已提交】最后一次成功写入的 ts
	lastRiseValue   float64         // 上次涨幅
	thisMinTs       int64           // 当前分钟的时间戳
	last2MinClose_8 uint64          // 最近两分钟的收盘价
	last1MinClose_8 uint64          // 最近一分钟的收盘价
	thisMinClose_8  uint64          // 当前分钟的收盘价
}

type cache_line_3 struct {
	StMeta          *symbolStatic.StaticTrade // 交易对静态信息
	bidPrice        atomic.Value              // 买一价,平仓和计算仓位价值用到
	takeProfitPrice float64                   // 止盈价格
	symbolIndex     systemx.SymbolIndex16I    // 交易对下标
	pScale          int32                     // 价格小数位
	qScale          int32                     // 数量小数位
	hasAllFilled    atomic.Bool               // 是否已经完全成交
	isStopLossAble  atomic.Bool               // 能否开始移动止损
	hasReceiveStop  bool                      // 是否已经收到过停止信号
	hasTreeNews     bool                      // 是否已经接受到treeNews
}

type cache_line_4 struct {
	clientOrderIdSmall string
	orderNum           decimal.Decimal          // 预挂单订单数量
	smallPercent       decimal.Decimal          // 小订单比例
	chanMarkPrice      chan toUpbitListChan.Job // 标记价格chan
	preAccountKeyId    uint8                    // 预挂单的账户id
	hasInit            bool                     // 是否已经预挂单初始化
}
type cache_line_5 struct {
	agLatencyTotal latency.Latency          // 延迟统计
	chanBookTick   chan toUpbitListChan.Job // 盘口数据chan
	seq            int64                    // 成功写入次数(严格递增)
}

type cache_line_6 struct {
	trigPriceMax_10 myMap.MySyncMap[int64, uint64] // 已触发品种的买入上限,秒级别时间戳和价格
}

type Single struct {
	cache_line_1
	cache_line_2
	cache_line_3
	cache_line_4
	cache_line_5
	cache_line_6
	chanAggTrade       chan toUpbitListChan.Job               // 成交数据chan
	chanTradeLite      chan []byte                            // 简易成交数据chan
	chanOrderUpdatePre chan []byte                            // delta订单数据chan
	chanMonitor        chan []byte                            // 订单监测chan
	chanOutSideSig     chan toUpbitListChan.Special           // 外部信号chan
	chanSuOrder        chan toUpBitListDataAfter.OnSuccessEvt // 成功订单chan
	secondArr          [11]*secondPerInfo                     // 每秒信息
	clientOrderIds     myMap.MySyncMap[string, uint8]         // 挂单成功的clientOrderId
	posTotalNeed       decimal.Decimal                        // 需要开仓的数量
	firstPriceBuy      decimal.Decimal                        // 当前应该下单的价格
	maxNotional        decimal.Decimal                        // 单品种最大开仓上限
	ctxStop            context.Context                        // 同步关闭ctx
	cancel             context.CancelFunc                     // 关闭函数
	pos                *toUpbitListPos.PosCal                 // 持仓计算对象
	twapSec            float64                                // twap下单间隔秒数
	globalStopLoss     float64                                // 全局止损价格
	closeDuration      time.Duration                          // 平仓持续时间
	thisOrderAccountId atomic.Int32                           // 当前订单使用的资金账户ID
	toAccountId        atomic.Int32                           // 准备接收资金的账户id
	symbolLen          uint16
}

func (s *Single) Clear() {
	s.newPriceMinWindowU64(toUpBitListDataStatic.TickCap)
	s.lastRiseValue = 0.0
}

func (s *Single) Start(accountKeyId uint8, index int, symbolName string) error {
	s.symbolIndex = systemx.SymbolIndex16I(index)
	s.symbolLen = uint16(len(symbolName))
	s.preAccountKeyId = accountKeyId
	symbolId, ok := symbolStatic.GetSymbol().GetSymbol(symbolName)
	if !ok {
		toUpBitListDataStatic.DyLog.GetLog().Errorf("symbolId not found: %s", symbolName)
		return toUpBitListDataStatic.ExType.GetNotSupportError("symbolId not found")
	}
	symbolKeyId := symbolInfo.MakeSymbolKey3(toUpBitListDataStatic.ExType, toUpBitListDataStatic.AcType, symbolId)
	// 初始化品种静态数据
	stMeta, err := symbolStatic.GetTrade().Get(symbolKeyId)
	if err != nil {
		toUpBitListDataStatic.DyLog.GetLog().Errorf("symbolKeyId %d not found", symbolKeyId)
		return err
	}
	s.StMeta = &stMeta
	// 初始化品种动态数据
	dyMeta, err := symbolDynamic.GetManager().Get(symbolKeyId)
	if err != nil {
		toUpBitListDataStatic.DyLog.GetLog().Errorf("symbolKeyId %d not found", symbolKeyId)
		return err
	}
	s.pScale = dyMeta.PScale
	s.qScale = dyMeta.QScale
	switch toUpBitListDataStatic.ExType {
	case exchangeEnum.BINANCE:
		// 1.15-->115
		s.upLimitPercent_2 = convertx.PriceStringToUint64(dyMeta.UpLimitPercent.String(), upPercentZoomScale)
	case exchangeEnum.BYBIT:
		// 0.15-->115
		s.upLimitPercent_2 = 100 + convertx.PriceStringToUint64(dyMeta.UpLimitPercent.String(), upPercentZoomScale)
	}
	s.newPriceMinWindowU64(toUpBitListDataStatic.TickCap)
	s.chanBookTick = make(chan toUpbitListChan.Job, 100)
	s.chanAggTrade = make(chan toUpbitListChan.Job, 10)
	s.chanMarkPrice = make(chan toUpbitListChan.Job, 10)
	s.chanTradeLite = make(chan []byte, 10)
	s.chanOrderUpdatePre = make(chan []byte, 10)
	s.chanMonitor = make(chan []byte, 10)
	s.chanOutSideSig = make(chan toUpbitListChan.Special, 10)
	s.chanSuOrder = make(chan toUpBitListDataAfter.OnSuccessEvt, 10)
	toUpbitListChan.RegisterMarket(index, s.chanBookTick, s.chanAggTrade, s.chanMarkPrice)
	toUpbitListChan.RegisterPrivate(index, s.chanTradeLite, s.chanOrderUpdatePre, s.chanMonitor, s.chanOutSideSig, s.chanSuOrder)
	latencyPrefix := idGen.BuildName2("GO", conf.ServerName)
	s.agLatencyTotal = latency.NewHttpMonitor(idGen.BuildName2(latencyPrefix, total), latency.PROCESS_TOTAL, resourceEnum.AGG_TRADE)
	s.btLatencyTotal = latency.NewHttpMonitor(idGen.BuildName2(latencyPrefix, total), latency.PROCESS_TOTAL, resourceEnum.BOOK_TICK)
	s.mpLatencyTotal = latency.NewHttpMonitor(idGen.BuildName2(latencyPrefix, total), latency.PROCESS_TOTAL, resourceEnum.MARK_PRICE)
	toUpBitListDataStatic.SymbolIndex.Store(symbolName, index)
	for i := range 11 {
		temp := &secondPerInfo{}
		temp.clear()
		s.secondArr[i] = temp
	}
	s.trigPriceMax_10 = myMap.NewMySyncMap[int64, uint64]()
	safex.SafeGo(symbolName+"单品种协程", s.onLoop)
	return nil
}

func (s *Single) onLoop() {
	for {
		select {
		case job := <-s.chanMarkPrice:
			s.onMarkPrice(job.Len, job.Buf)
		case job := <-s.chanAggTrade:
			s.onAggTrade(job.Len, job.Buf)
		case job := <-s.chanBookTick:
			s.onBookTick(job.Len, job.Buf)
		case job := <-s.chanTradeLite:
			s.onTradeLite(job)
		case job := <-s.chanOrderUpdatePre:
			s.onPayloadOrder(job)
		case job := <-s.chanMonitor:
			s.onMonitorData(job)
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
