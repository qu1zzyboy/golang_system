package bnDrive_upbitKrw

import (
	"context"
	"lowLatencyServer/internal/strategy/newsDrive/bn/bnDriveCommon"
	"lowLatencyServer/internal/strategy/newsDrive/common/driverStatic"
	"time"
	"upbitBnServer/internal/infra/safex"
)

func (s *Single) tryBuyLoop(ctxGlobal context.Context, max int32) {
	//开启每秒抢一次的协程,来抢未来十秒的订单
	safex.SafeGo("to_upbit_bn_open_second", func() {
		var i int32
		defer func() {
			driverStatic.DyLog.GetLog().Infof("每秒抽奖协程结束,抽奖次数[当前抽奖序号:%d,max:%d]", i, max)
		}()
		// 从账户3开始
		for i = 3; i < max; i++ {
			select {
			case <-ctxGlobal.Done():
				driverStatic.DyLog.GetLog().Infof("收到关闭信号,退出每秒抽奖协程")
				return
			default:
				// 睡到这一秒的965毫秒
				now := time.Now()
				secStart := now.Truncate(time.Second)
				target := secStart.Add(965 * time.Millisecond)

				// 如果已经超过 965ms，就睡到下一秒的 965ms
				if !now.Before(target) {
					i++
					target = target.Add(time.Second)
				}
				time.Sleep(time.Until(target))

				ctxThis, cancel := context.WithCancel(context.Background())    // 每秒的停止信号
				placeIndex := uint8(bnDriveCommon.GetCurIndex(i))              // 该秒的下单账户id
				fromAccountId := bnDriveCommon.GetPreIndex(i)                  // 该秒的撤单账户id
				s.order.RefreshSecond(cancel, s.pos.GetThisOpen(), placeIndex) // 刷新每秒开始

				driverStatic.DyLog.GetLog().Infof("==========[循环序号:%d,下单账户:%d,撤单账户:%d]秒下单=========", i, placeIndex, fromAccountId)

				// 撤销上一轮的订单
				go s.pos.CancelAndTransfer(i, fromAccountId, s.symbol.Sym.SymbolName)

				//探测逻辑
				go s.order.MonitorPerNormal(ctxGlobal, ctxThis, placeIndex)

				//真实下单逻辑
				go s.order.PlacePerNormal(ctxGlobal, ctxThis)

				go func() {
					time.Sleep(time.Second)
					cancel()
				}()
			}
		}
	})
}
