package toUpBitListBnExecute

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/hhh500/quantGoInfra/infra/safex"
	"github.com/hhh500/quantGoInfra/pkg/container/map/myMap"
	"github.com/hhh500/quantGoInfra/pkg/singleton"
	"github.com/hhh500/upbitBnServer/internal/quant/execute"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderModel"
	"github.com/hhh500/upbitBnServer/internal/quant/market/symbolInfo/symbolStatic"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpbitListPos"
	"github.com/shopspring/decimal"
)

type StopType uint8

const (
	StopByTreeNews StopType = iota
	StopByMoveStopLoss
	StopByBtTakeProfit
	StopByGetCmcFailure
	StopByGetRemoteFailure
)

const (
	order_from = orderBelongEnum.TO_UPBIT_LIST_LOOP
)

var (
	serviceSingleton = singleton.NewSingleton(func() *Execute {
		return &Execute{
			clientOrderIds: myMap.NewMySyncMap[string, uint8](),
			pos:            toUpbitListPos.NewPosCal(),
		}
	})
	stopReasonArr = []string{
		"未触发TreeNews",
		"%5移动止损触发",
		"BookTick止盈触发",
		"获取cmc_id失败",
		"获取远程参数失败",
	}
)

func GetExecute() *Execute {
	return serviceSingleton.Get()
}

type Execute struct {
	closeDecArr             [11]decimal.Decimal            // 每个账户每秒应该止盈的数量
	maxNotionalArr          [11]atomic.Value               // 每个账户的最大开仓上限
	clientOrderIds          myMap.MySyncMap[string, uint8] // 挂单成功的clientOrderId
	stopThisSecondPerArr    [11]atomic.Bool                // 本轮是否可以参与抽奖,挂单成功即为true
	hasInToSecondPerLoopArr [11]atomic.Bool                // 是否已经进入了每秒抽奖循环
	posTotalNeed            decimal.Decimal                // 需要开仓的数量
	firstPriceBuy           decimal.Decimal                // 当前应该下单的价格
	maxNotional             decimal.Decimal                // 单品种最大开仓上限
	ctxStop                 context.Context                // 同步关闭ctx
	cancel                  context.CancelFunc             // 关闭函数
	pos                     *toUpbitListPos.PosCal         // 持仓计算对象
	StMeta                  *symbolStatic.StaticTrade      // 交易规范
	takeProfitPrice         float64                        // 止盈价格
	twapSec                 float64                        // twap下单间隔秒数
	symbolIndex             int                            // 触发的交易对索引
	closeDuration           time.Duration                  // 平仓持续时间
	hasAllFilled            atomic.Bool                    // 是否已经完全成交
	hasTreeNews             atomic.Bool                    // 是否接收到过TreeNews
	hasReceiveStop          atomic.Bool                    // 是否已经收到过停止信号
	thisOrderAccountId      atomic.Int32                   // 当前订单使用的资金账户ID
	toAccountId             atomic.Int32                   // 准备接收资金的账户id
	PScale                  int32                          // 价格小数位
	QScale                  int32                          // 数量小数位

}

func (s *Execute) ReceiveTreeNews(symbolIndex int) {
	if symbolIndex == s.symbolIndex {
		s.hasTreeNews.Store(true)
	}
}

func (s *Execute) clear() {
	s.posTotalNeed = decimal.Zero
	s.pos.Clear() //清空持仓统计
	s.PScale = 0
	s.QScale = 0
	s.StMeta = nil
	s.takeProfitPrice = 0
	for i := range s.stopThisSecondPerArr {
		s.hasInToSecondPerLoopArr[i].Store(false)
		s.stopThisSecondPerArr[i].Store(false)
		s.closeDecArr[i] = decimal.Zero
	}
	s.hasAllFilled.Store(false)
	s.thisOrderAccountId.Store(0)
	toUpBitListDataAfter.ClearTrig()
}

func (s *Execute) StartTrig(trigSymbolIndex int, pScale, qScale int32, stMeta *symbolStatic.StaticTrade) {
	s.symbolIndex = trigSymbolIndex
	s.PScale = pScale
	s.QScale = qScale
	s.StMeta = stMeta
	s.ctxStop, s.cancel = context.WithCancel(context.Background())
	var ok bool
	s.maxNotional, ok = toUpBitListDataStatic.SymbolMaxNotional.Load(trigSymbolIndex)
	if !ok {
		s.maxNotional = decimal.NewFromInt(50000)
	}
	s.maxNotional = s.maxNotional.Sub(toUpBitListDataStatic.Dec500)
}

func (s *Execute) SetExecuteParam(trigPrice float64, twapSec float64) {
	s.twapSec = twapSec
	s.takeProfitPrice = trigPrice
	s.closeDuration = time.Duration(twapSec) * time.Second
	toUpBitListDataStatic.DyLog.GetLog().Infof("止盈价格: %.8f,平仓持续时间: %s,单账户上限:%s", trigPrice, s.closeDuration.String(), s.maxNotional)
}

func (s *Execute) ReceiveStop(stopType StopType) {
	if s.hasReceiveStop.Load() {
		return
	}
	s.hasReceiveStop.Store(true)
	toUpBitListDataStatic.DyLog.GetLog().Infof("收到停止信号: %s", stopReasonArr[stopType])
	s.cancel()
	//开启平仓线程
	safex.SafeGo("toUpBitListBn_close", func() {
		defer func() {
			toUpBitListDataStatic.DyLog.GetLog().Infof("当前账户id[%d] 平仓协程结束", s.thisOrderAccountId.Load())
			time.Sleep(20 * time.Millisecond)
			s.clear()
		}()
		// 撤销全部订单
		s.clientOrderIds.Range(func(clientOrderId string, accountKeyId uint8) bool {
			bnOrderAppManager.GetTradeManager().SendCancelOrder(order_from, accountKeyId, &orderModel.MyQueryOrderReq{
				ClientOrderId: clientOrderId,
				StaticMeta:    s.StMeta,
			})
			return true
		})

		// 判断有没有持仓
		use := s.pos.GetTotal()
		if use.LessThanOrEqual(decimal.Zero) {
			toUpBitListDataStatic.DyLog.GetLog().Infof("没有可用的平仓数量,取消平仓")
			return
		}
		if use.LessThanOrEqual(toUpBitListDataStatic.Dec500) {
			toUpBitListDataStatic.DyLog.GetLog().Infof("没有足够的平仓数量,取消平仓")
			return
		}
		//每秒平一次
		perDec := decimal.NewFromFloat(1 / s.twapSec)
		copyMap := s.pos.GetAllAccountPos()
		for accountKeyId, vol := range copyMap {
			s.closeDecArr[accountKeyId] = vol.Mul(perDec).Truncate(s.QScale) //每秒应该止盈的数量
		}
		ticker := time.NewTicker(time.Second)
		timeout := time.After(s.closeDuration)
		for {
			select {
			case <-ticker.C:
				{
					lastBid64, ok := toUpBitListDataAfter.LoadBidPrice()
					if !ok {
						continue
					}
					priceDec := decimal.NewFromFloat(lastBid64).Truncate(s.PScale)
					posLeft := s.pos.GetTotal()
					if s.pos.GetTotal().Mul(priceDec).LessThanOrEqual(toUpBitListDataStatic.Dec500) {
						toUpBitListDataStatic.DyLog.GetLog().Infof("平仓完全成交,开始清理资源")
						ticker.Stop()
						return
					}
					toUpBitListDataStatic.DyLog.GetLog().Infof("============开始平仓,剩余:%s============", posLeft)
					// 最新的每个账户的仓位情况
					copyMap := s.pos.GetAllAccountPos()
					for accountKeyId, vol := range copyMap {
						// 已经完全平完了
						if vol.LessThanOrEqual(decimal.Zero) {
							continue
						}
						// 不够就全平
						num := s.closeDecArr[accountKeyId]
						if vol.LessThan(num) {
							num = vol.Truncate(s.QScale)
						}
						// 发送平仓信号
						if err := bnOrderAppManager.GetTradeManager().SendPlaceOrder(order_from, accountKeyId, s.symbolIndex,
							&orderModel.MyPlaceOrderReq{
								OrigPrice:     priceDec,
								OrigVol:       num,
								ClientOrderId: toUpBitListDataStatic.GetClientOrderIdBy("close"),
								StaticMeta:    s.StMeta,
								OrderType:     execute.ORDER_TYPE_LIMIT,
								OrderMode:     execute.ORDER_SELL_CLOSE,
							}); err != nil {
							toUpBitListDataStatic.DyLog.GetLog().Errorf("每秒平仓创建订单失败: %v", err)
						}
					}
				}
			case <-timeout:
				toUpBitListDataStatic.DyLog.GetLog().Infof("平仓时间结束,开始清理资源")
				ticker.Stop()
				return
			}
		}
	})
}
