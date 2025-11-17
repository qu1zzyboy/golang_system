package toUpbitBybitSymbol

import (
	"time"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
)

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

func (s *Single) tryBuyLoop(max int32) {
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
				placeIndex := uint8(getCurIndex(i)) // 该秒的下单账户id
				fromAccountId := getPreIndex(i)     // 该秒的撤单账户id

				dynamicLog.Log.GetLog().Infof("==========[循环序号:%d,下单账户:%d,撤单账户:%d]秒下单=========", i, placeIndex, fromAccountId)

				//真实下单逻辑
				go s.placePer(i, placeIndex)
			}
		}
	})
}
