package toUpbitListBnSymbol

import (
	"time"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
)

func (s *Single) tryBuyLoopBithumbKrw(max int32) {
	//开启每秒抢一次的协程,来抢未来十秒的订单
	safex.SafeGo("to_upbit_bn_open_second", func() {
		var i int32
		defer func() {
			toUpBitDataStatic.DyLog.GetLog().Infof("每秒抽奖协程结束,抽奖次数[当前抽奖序号:%d,max:%d]", i, max)
		}()
		for i = 3; i < max; i++ {
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

				// 如果已经超过 965ms，就睡到下一秒的 965ms
				if !now.Before(target) {
					target = target.Add(time.Second)
				}
				time.Sleep(time.Until(target))

				//已经完全开满
				if s.hasAllFilled.Load() {
					break
				}
				// 进入每秒抽奖循环
				placeIndex := uint8(getCurIndex(i))           // 该秒的下单账户id
				s.SecondArr[placeIndex].start()               // 重置该秒状态
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
