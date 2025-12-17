package driverOrderDedup

import (
	"sync/atomic"
	"time"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/quant/execute/order/bnOrderDedup"
	"upbitBnServer/internal/quant/execute/order/orderStaticMeta"
	"upbitBnServer/internal/quant/execute/plan/cancelPlan"
	"upbitBnServer/internal/quant/execute/plan/placePlan"
	"upbitBnServer/internal/quant/execute/plan/updatePlan"
	"upbitBnServer/internal/strategy/newsDrive/common/driverStatic"
	"upbitBnServer/pkg/container/hashTable"
	"upbitBnServer/pkg/container/map/myMap"
)

// 负载因子 <= 0.5 时性能最稳
// capacity >= 每秒订单数 × 2*2  //ws条数 *0.5的负载因子
// cap >= 800 × 2 = 1600

var (
	Upbit_pre     atomic.Int64 //多线程访问，可能会同时订阅两份payload数据
	Upbit_main    atomic.Int64
	Deduper       = hashTable.NewSlidingWindow(2, 256, bnOrderDedup.HashOrderKey)
	PlaceManager  = myMap.NewMySyncMap[systemx.WsId16B, *placePlan.PlacePlan]()   //clientOrderId-->下单计划
	ModifyManager = myMap.NewMySyncMap[systemx.WsId16B, *updatePlan.UpdatePlan]() //clientOrderId-->改单计划
	CancelManager = myMap.NewMySyncMap[systemx.WsId16B, *cancelPlan.CancelPlan]() //clientOrderId-->改单计划
)

/***该策略全局的交易计划***/

func DelOrderStatic(clientOrderId systemx.WsId16B) {
	go func() {
		time.Sleep(2 * time.Second)
		orderStaticMeta.GetService().DelOrderMeta(clientOrderId) // 删除成交的订单信息
		driverStatic.DyLog.GetLog().Infof("当前系统内还有:%d 静态订单数据", orderStaticMeta.GetService().GetLength())
	}()
}

func StartPlanCheck() {
	safex.SafeGo("plan_check", func() {
		ticker_1 := time.NewTicker(1 * time.Second)
		ticker_5 := time.NewTicker(5 * time.Second)
		const limit = 5000

		for {
			select {
			case <-ticker_1.C:
				Deduper.Rotate()
			case <-ticker_5.C:
				ts := time.Now().UnixMilli()
				PlaceManager.Range(func(clientOrderId systemx.WsId16B, v *placePlan.PlacePlan) bool {
					if ts-v.UpdateAt > limit {
						driverStatic.DyLog.GetLog().Errorf("下单计划[%s,%s]未接受到下单返回", v.Req.SymbolName, clientOrderId)
						PlaceManager.Delete(clientOrderId)
					}
					return true
				})
				ModifyManager.Range(func(k systemx.WsId16B, v *updatePlan.UpdatePlan) bool {
					if ts-v.UpdateAt > limit {
						driverStatic.DyLog.GetLog().Errorf("改单计划[%s,%s]未接受到改单返回", v.Req.SymbolName, k)
						ModifyManager.Delete(k)
					}
					return true
				})
				CancelManager.Range(func(k systemx.WsId16B, v *cancelPlan.CancelPlan) bool {
					if ts-v.UpdateAt > limit {
						driverStatic.DyLog.GetLog().Errorf("撤单计划[%s,%s]未接受到撤单返回", v.Req.SymbolName, k)
						CancelManager.Delete(k)
					}
					return true
				})
			}
		}
	})
}

func Is_UPBIT_PRE_TRY_TRIG(ts int64) bool {
	for {
		old := Upbit_pre.Load()
		if ts <= old {
			return false
		}
		if Upbit_pre.CompareAndSwap(old, ts) {
			return true
		}
	}
}

func Is_UPBIT_MAIN_New(ts int64) bool {
	for {
		old := Upbit_main.Load()
		if ts < old {
			return false
		}
		if Upbit_main.CompareAndSwap(old, ts) {
			return true
		}
	}
}
