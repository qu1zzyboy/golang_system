package toUpbitListBnSymbol

import (
	"strings"
	"time"

	"github.com/hhh500/quantGoInfra/conf"
	"github.com/hhh500/quantGoInfra/infra/safex"
	"github.com/hhh500/quantGoInfra/pkg/utils/convertx"
	"github.com/hhh500/quantGoInfra/pkg/utils/idGen"
	"github.com/hhh500/quantGoInfra/quant/exchanges/exchangeEnum"
	"github.com/hhh500/quantGoInfra/resource/resourceEnum"
	"github.com/hhh500/upbitBnServer/internal/infra/latency"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"github.com/hhh500/upbitBnServer/internal/quant/market/symbolInfo"
	"github.com/hhh500/upbitBnServer/internal/quant/market/symbolInfo/symbolDynamic"
	"github.com/hhh500/upbitBnServer/internal/quant/market/symbolInfo/symbolStatic"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpbitListChan"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

const (
	total                    = "BINANCE_TOTAL"
	jsonEvent                = "E"
	upPercentZoomScale int32 = 2 // 价格放大的倍数,10的ZoomScale次方倍
	ws_req_from              = orderBelongEnum.TO_UPBIT_LIST_PRE
)

type Single struct {
	trigHotData
	clientOrderIdSmall string
	mpLatencyTotal     latency.Latency           //延迟统计
	btLatencyTotal     latency.Latency           //延迟统计
	agLatencyTotal     latency.Latency           //延迟统计
	orderNum           decimal.Decimal           //预挂单订单数量
	smallPercent       decimal.Decimal           //小订单比例
	lastRiseValue      float64                   //上次涨幅
	chanMarkPrice      chan toUpbitListChan.Job  // 标记价格chan
	chanAggTrade       chan toUpbitListChan.Job  // 成交数据chan
	chanBookTick       chan toUpbitListChan.Job  // 盘口数据chan
	chanTradeLite      chan []byte               // 简易成交数据chan
	chanDeltaOrder     chan []byte               // delta订单数据chan
	chanMonitor        chan []byte               // 订单监测chan
	thisMinTs          int64                     // 当前分钟的时间戳
	priceMax_10        uint64                    // 价格上限(放大了10的10次方倍)
	lastMarkPrice_8    uint64                    // 上次标记价格
	markPrice_8        uint64                    // 最新标记价格(放大了10的8次方倍)
	last2MinClose_8    uint64                    // 最近两分钟的收盘价
	last1MinClose_8    uint64                    // 最近一分钟的收盘价
	thisMinClose_8     uint64                    // 当前分钟的收盘价
	upLimitPercent_2   uint64                    // 涨停百分比,115就是1.15
	symbolIndex        int                       // 交易对下标
	StMeta             *symbolStatic.StaticTrade // 交易对静态信息
	pScale             int32                     // 价格小数位
	qScale             int32                     // 数量小数位
	accountKeyId       uint8                     // 预挂单的账户id
	hasInit            bool
}

func (s *Single) Clear() {
	s.newPriceMinWindowU64(toUpBitListDataStatic.TickCap)
	s.lastRiseValue = 0.0
}

func (s *Single) Start(accountKeyId uint8, index int, symbolName string) error {
	s.accountKeyId = accountKeyId
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
	s.chanDeltaOrder = make(chan []byte, 10)
	s.chanMonitor = make(chan []byte, 10)
	toUpbitListChan.RegisterMarket(index, s.chanBookTick, s.chanAggTrade, s.chanMarkPrice)
	toUpbitListChan.RegisterPrivate(index, s.chanTradeLite, s.chanDeltaOrder, s.chanMonitor)
	latencyPrefix := idGen.BuildName2("GO", conf.ServerName)
	s.agLatencyTotal = latency.NewHttpMonitor(idGen.BuildName2(latencyPrefix, total), latency.PROCESS_TOTAL, resourceEnum.AGG_TRADE)
	s.btLatencyTotal = latency.NewHttpMonitor(idGen.BuildName2(latencyPrefix, total), latency.PROCESS_TOTAL, resourceEnum.BOOK_TICK)
	s.mpLatencyTotal = latency.NewHttpMonitor(idGen.BuildName2(latencyPrefix, total), latency.PROCESS_TOTAL, resourceEnum.MARK_PRICE)
	s.symbolIndex = index
	toUpBitListDataStatic.SymbolIndex.Store(symbolName, index)
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
		case job := <-s.chanDeltaOrder:
			s.onPayloadOrder(job)
		case job := <-s.chanMonitor:
			s.onMonitorData(job)
		}
	}
}

func (s *Single) onMonitorData(data []byte) {
	// s := "Limit price can't be higher than 4550.62."
	errMsg := gjson.GetBytes(data, "error.msg").String()
	parts := strings.Fields(errMsg)      // 按空格切分
	last := parts[len(parts)-1]          // 最后一段 "4550.62."
	last = strings.TrimSuffix(last, ".") // 去掉末尾的点
	// 价格更新,放大10的10次方倍
	monitor_10 := convertx.PriceStringToUint64(last, 10)
	if monitor_10 == s.priceMax_10 {
		return
	}
	s.priceMax_10 = monitor_10
	toUpBitListDataAfter.TrigPriceMax_10.Store(time.Now().Unix(), monitor_10)
	toUpBitListDataStatic.DyLog.GetLog().Infof("最新探测价格[%d]: %s", monitor_10, errMsg)
}

// {
//     "id": "Ptest123456",
//     "status": 400,
//     "error": {
//         "code": -4016,
//         "msg": "Limit price can't be higher than 4275.37."
//     }
// }
