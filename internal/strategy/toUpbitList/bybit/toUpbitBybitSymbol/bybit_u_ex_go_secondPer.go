package toUpbitBybitSymbol

import (
	"time"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
)

func (s *Single) tryBuyLoop(max int32) {
	//开启每秒抢一次的协程,来抢未来十秒的订单
	safex.SafeGo("to_upbit_bybit_open_second", func() {
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
				next := now.Truncate(time.Second).Add(time.Second)
				trigger := next.Add(+5 * time.Millisecond)
				sleep := time.Until(trigger)
				if sleep > 0 {
					time.Sleep(sleep)
				}
				//已经完全开满
				if s.hasAllFilled.Load() {
					break
				}
				// 进入每秒抽奖循环
				dynamicLog.Log.GetLog().Infof("==========[循环序号:%d]秒下单=========", i)
				//真实下单逻辑
				go s.placePer()
			}
		}
	})
}
