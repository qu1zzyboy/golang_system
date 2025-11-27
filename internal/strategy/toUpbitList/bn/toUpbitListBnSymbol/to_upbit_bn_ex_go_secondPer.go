package toUpbitListBnSymbol

import (
	"sync/atomic"
	"time"

	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"

	"github.com/shopspring/decimal"
)

var (
	trans = [11]int32{
		3,  //账户0-->账户3
		4,  //账户1-->账户4
		5,  //账户2-->账户5
		6,  //账户3-->账户6
		7,  //账户4-->账户7
		8,  //账户5-->账户8
		9,  //账户6-->账户9
		10, //账户7-->账户10
		1,  //账户8-->账户1
		2,  //账户9-->账户2
		3,  //账户10-->账户3
	}
)

type secondPerInfo struct {
	maxNotional          atomic.Value
	hasInToSecondPerLoop atomic.Bool
	stopThisSecondPer    atomic.Bool
}

func (s *secondPerInfo) clear() {
	s.maxNotional.Store(decimal.Zero)
	s.hasInToSecondPerLoop.Store(false)
	s.stopThisSecondPer.Store(false)
}

func (s *secondPerInfo) start() {
	s.stopThisSecondPer.Store(false)   // 开启本轮抽奖信号
	s.hasInToSecondPerLoop.Store(true) // 确认进入了每秒抽奖循环
}

func (s *secondPerInfo) receiveStop(accountKeyId uint8) {
	if s.hasInToSecondPerLoop.Load() {
		if !s.stopThisSecondPer.Load() {
			s.stopThisSecondPer.Store(true)
			toUpBitDataStatic.DyLog.GetLog().Infof("账户[%d]没钱,停止这一秒抽奖", accountKeyId)
		}
	}
}

func (s *secondPerInfo) loadStop() bool {
	return s.stopThisSecondPer.Load()
}

// 给定一个整数 i,算出它在 1–10 的“循环序列”里的前2个下标
func getPreIndex(i int32) int32 {
	res := i%10 + 8
	if res > 10 {
		res = res - 10
	}
	return res
}

// 把任意整数 i 映射到 1–10 之间
func getCurIndex(i int32) int32 {
	res := i % 10
	if res == 0 {
		res = 10
	}
	return res
}

func (s *Single) TryBuyLoop(max int32) {
	//开启每秒抢一次的协程,来抢未来十秒的订单
	safex.SafeGo("to_upbit_bn_open_second", func() {
		var i int32
		defer func() {
			toUpBitDataStatic.DyLog.GetLog().Infof("每秒抽奖协程结束,抽奖次数[当前抽奖序号:%d,max:%d]", i, max)
		}()
		for i = 1; i < max; i++ {
			if i >= 4 {
				s.isStopLossAble.Store(true)
			}
			select {
			case <-s.ctxStop.Done():
				toUpBitDataStatic.DyLog.GetLog().Infof("收到关闭信号,退出每秒抽奖协程")
				return
			default:
				// 睡到下一秒的5毫秒后
				now := time.Now()
				secStart := now.Truncate(time.Second)
				target := secStart.Add(965 * time.Millisecond)

				if now.Before(target) {
					time.Sleep(time.Until(target))
				}

				// next := now.Truncate(time.Second).Add(time.Second)
				// trigger := next.Add(+5 * time.Millisecond)
				// sleep := time.Until(target)
				// if sleep > 0 {
				// 	time.Sleep(sleep)
				// }
				//已经完全开满
				if s.hasAllFilled.Load() {
					break
				}
				// 进入每秒抽奖循环
				placeIndex := uint8(getCurIndex(i))           // 该秒的下单账户id
				s.secondArr[placeIndex].start()               // 重置该秒状态
				s.thisOrderAccountId.Store(int32(placeIndex)) // 当前订单使用的资金账户Id
				fromAccountId := getPreIndex(i)               // 该秒的撤单账户id
				s.toAccountId.Store(trans[fromAccountId])     // 当前应该接收资金的账户,新的一秒开始就更新

				dynamicLog.Log.GetLog().Infof("==========[循环序号:%d,下单账户:%d,撤单账户:%d]秒下单=========", i, placeIndex, fromAccountId)

				// 撤销上一轮的订单
				go s.cancelAndTransfer(i, fromAccountId)

				//探测逻辑
				go s.monitorPer(placeIndex)

				//真实下单逻辑
				go s.placePer(i, placeIndex)
			}
		}
	})
}
