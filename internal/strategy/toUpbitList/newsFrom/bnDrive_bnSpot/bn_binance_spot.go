package bnDrive_bnSpot

import (
	"lowLatencyServer/internal/quant/execute/order/bnOrderWrap"
	"lowLatencyServer/internal/strategy/newsDrive/bn/bnDrivePos"
	"lowLatencyServer/internal/strategy/newsDrive/common/driverSymbol"
	"lowLatencyServer/server/instanceEnum"
	"lowLatencyServer/server/usageEnum"
	"ofeisInfra/infra/safex"
	"ofeisInfra/infra/systemx"
	"ofeisInfra/quant/execute"
	"privateInfra/pkg/utils/time2str"
	"privateInfra/pkg/utils/u64Cal"
	"privateInfra/quant/execute/order/orderModel"
	"time"
)

const (
	to_upbit_main = usageEnum.NEWS_DRIVE_MAIN
	from_bn       = instanceEnum.DRIVER_LIST_BN
	f_per         = 25.0
	total         = 25
)

var (
	clientOrderArr [total]systemx.WsId16B
)

type Single struct {
	symbol          *driverSymbol.Symbol     // 交易对信息
	pos             *bnDrivePos.CacheLinePos // 仓位对象
	perNum          systemx.OrderSdkType     // 每一个订单需要放置的订单数量
	takeProfitPrice float64                  // 止盈价格
	accountKeyId    uint8
	hasReceiveStop  bool // 是否已经收到过停止信号
}

func NewSingle(symbol *driverSymbol.Symbol, pos *bnDrivePos.CacheLinePos, takeProfitPrice float64, accountKeyId uint8) *Single {
	return &Single{symbol: symbol, pos: pos, takeProfitPrice: takeProfitPrice, accountKeyId: accountKeyId}
}

func (s *Single) Start(f14, f15, f16 float64) {
	safex.SafeGo("bnSpot_initBuyOpen", func() {
		s.initBuyOpen(f14, f15, f16)
		s.buyLoop()
	})
}

func (s *Single) initBuyOpen(f14, f15, f16 float64) {
	pScale := s.symbol.Sym.PScale
	qScale := s.symbol.Sym.QScale
	_, posLeft := s.pos.GetAllAccountPos()
	perAvg := posLeft / f_per
	s.perNum = u64Cal.FromF64(perAvg, qScale.Uint8())
	var req = &orderModel.MyPlaceOrderReq{
		SymbolName:  s.symbol.Sym.SymbolName,
		Qvalue:      s.perNum,
		Pscale:      pScale,
		Qscale:      qScale,
		OrderMode:   execute.BUY_OPEN_LIMIT,
		SymbolIndex: s.symbol.Sym.SymbolIndex,
		ReqFrom:     from_bn,
		UsageFrom:   to_upbit_main,
	}
	for i := range 25 {
		clientOrderId := time2str.GetNowTimeStampMicroSlice16()
		req.ClientOrderId = clientOrderId
		clientOrderArr[i] = clientOrderId
		switch i % 3 {
		case 0:
			req.Pvalue = u64Cal.FromF64(f14, pScale.Uint8())
			bnOrderWrap.PlaceOrderWithPlan(s.accountKeyId, req)
		case 1:
			req.Pvalue = u64Cal.FromF64(f15, pScale.Uint8())
			bnOrderWrap.PlaceOrderWithPlan(s.accountKeyId, req)
		case 2:
			req.Pvalue = u64Cal.FromF64(f16, pScale.Uint8())
			bnOrderWrap.PlaceOrderWithPlan(s.accountKeyId, req)
		}
	}
}

func (s *Single) buyLoop() {
	pScale := s.symbol.Sym.PScale
	qScale := s.symbol.Sym.QScale
	var i = 0
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		i++
		if i >= total {
			ticker.Stop()
			return
		}
		val := s.symbol.Sym.BidPrice.Load()
		if val == nil {
			continue
		}
		bid64 := val.(float64)
		bnOrderWrap.ModifyOrderWithPlan(s.accountKeyId, &orderModel.MyModifyOrderReq{
			SymbolName:    s.symbol.Sym.SymbolName,
			ClientOrderId: clientOrderArr[i],
			Pvalue:        u64Cal.FromF64(bid64, pScale.Uint8()) + 5,
			Qvalue:        s.perNum,
			Pscale:        pScale,
			Qscale:        qScale,
			OrderMode:     execute.BUY_OPEN_LIMIT,
			ReqFrom:       from_bn,
			UsageFrom:     to_upbit_main,
		})
	}
}

func (s *Single) sellLoop() {
	if s.hasReceiveStop {
		return
	}
	s.hasReceiveStop = true

	pScale := s.symbol.Sym.PScale
	qScale := s.symbol.Sym.QScale
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		val := s.symbol.Sym.BidPrice.Load()
		if val == nil {
			continue
		}
		bid64 := val.(float64)
		bnOrderWrap.PlaceOrderWithPlan(s.accountKeyId, &orderModel.MyPlaceOrderReq{
			SymbolName:    s.symbol.Sym.SymbolName,
			ClientOrderId: time2str.GetNowTimeStampMicroSlice16(),
			Pvalue:        u64Cal.FromF64(bid64, pScale.Uint8()) - 2,
			Qvalue:        s.perNum,
			Pscale:        pScale,
			Qscale:        qScale,
			OrderMode:     execute.SELL_CLOSE_LIMIT,
			SymbolIndex:   s.symbol.Sym.SymbolIndex,
			ReqFrom:       from_bn,
			UsageFrom:     to_upbit_main,
		})
	}
}
