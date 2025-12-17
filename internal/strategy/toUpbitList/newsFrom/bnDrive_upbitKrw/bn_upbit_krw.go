package bnDrive_upbitKrw

import (
	"ofeisInfra/infra/safex"
	"time"
)

type Single struct {
	symbol          *driverSymbol.Symbol // 交易对信息
	order           *bnDriveOrder.CacheLineOrder
	pos             *bnDrivePos.CacheLinePos
	takeProfitPrice float64 // 止盈价格
	hasReceiveStop  bool    // 是否已经收到过停止信号
}

func NewSingle(order *bnDriveOrder.CacheLineOrder, pos *bnDrivePos.CacheLinePos, symbol *driverSymbol.Symbol) *Single {
	return &Single{symbol: symbol, order: order, pos: pos}
}

func (s *Single) Start() {
	s.tryBuyLoop(s.pos.GetGlobal(), toUpbitParam.MaxAccount)
}

// driverDefine.StopByBtTakeProfit 触发止盈价格停止
func (s *Single) receiveStop(stopType driverDefine.StopType) {
	if s.hasReceiveStop {
		return
	}
	s.hasReceiveStop = true
	driverStatic.DyLog.GetLog().Infof("收到停止信号==> %s", driverDefine.StopReasonArr[stopType])
	s.pos.CancelGlobal()
	//开启平仓线程
	safex.SafeGo("bn_driver_upbit_close", func() {
		defer func() {
			driverStatic.DyLog.GetLog().Info("平仓协程结束")
			time.Sleep(2 * time.Second)
			s.symbol.Clear()
			s.order.Clear()
			s.pos.Clear()
			driverStatic.ClearTrig()
		}()
		// 撤销全部订单
		s.pos.CancelAllOrders(s.symbol.Sym.SymbolName)

		if !s.pos.IsNeedClose() {
			return
		}
		//每秒平一次
		var closeDecArr [toUpbitParam.MaxAccount]float64 // 每个账户每秒应该止盈的数量
		var copyMap map[uint8]float64
		var posLeft float64

		copyMap, posLeft = s.pos.GetAllAccountPos()
		for accountKeyId, vol := range copyMap {
			closeDecArr[accountKeyId] = 0.1 * vol //每秒应该止盈的数量
		}
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			val := s.symbol.Sym.BidPrice.Load()
			if val == nil {
				continue
			}
			bid64 := val.(float64)
			copyMap, posLeft = s.pos.GetAllAccountPos()
			if posLeft*bid64 <= toUpbitParam.Dec500 {
				driverStatic.DyLog.GetLog().Infof("平仓完全成交,开始清理资源")
				ticker.Stop()
				return
			}
			driverStatic.DyLog.GetLog().Infof("============开始平仓,剩余:%.8f============", posLeft)
			// 最新的每个账户的仓位情况
			for accountKeyId, vol := range copyMap {
				// 已经完全平完了
				if vol <= 0 {
					continue
				}
				// 不够就全平
				num := closeDecArr[accountKeyId]
				if vol < num {
					num = vol
				}
				s.order.CloseOrderNormal(bid64, num, accountKeyId)
			}
		}
	})
}
