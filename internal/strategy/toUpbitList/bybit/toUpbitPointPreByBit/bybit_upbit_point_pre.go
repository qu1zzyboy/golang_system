package toUpbitPointPreByBit

import (
	"math"
	"time"
	"upbitBnServer/internal/cal/u64Cal"
	"upbitBnServer/internal/infra/observe/notify/notifyTg"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/instanceEnum"
	"upbitBnServer/internal/infra/systemx/usageEnum"
	"upbitBnServer/internal/quant/exchanges/bybit/order/byBitOrderAppManager"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolDynamic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitPoint"
	"upbitBnServer/pkg/container/map/myMap"
	"upbitBnServer/pkg/utils/time2str"
)

const (
	point_pre  = usageEnum.TO_UPBIT_PRE
	from_bybit = instanceEnum.TO_UPBIT_LIST_BYBIT
)

var (
	clientOrderOnline  = myMap.NewMySyncMap[systemx.WsId16B, struct{}]() //clientOrderId-->占位符,所有的挂单状态的订单
	ClientOrderNotOpen = myMap.NewMySyncMap[systemx.WsId16B, struct{}]() //clientOrderId-->占位符,有就不下单
)

func OnOrderUpdate(isOnline bool, clientOrderId systemx.WsId16B) {
	// 挂单状态就存,非挂单状态就删
	if isOnline {
		clientOrderOnline.Store(clientOrderId, struct{}{})
	} else {
		clientOrderOnline.Delete(clientOrderId)
	}
}

type PointPre struct {
	clientOrderIdSmall systemx.WsId16B          // 探针id
	lastMarkPrice      float64                  // 上次标记价格,1倍
	PointNum           uint64                   // 预挂单订单数量,放大了10的qScale次方
	symbolKeyId        uint64                   //
	symbolLen          uint16                   // symbol长度
	symbolIndex        systemx.SymbolIndex16I   // 交易对下标
	preAccountKeyId    uint8                    // 预挂单的账户id
	upT                toUpbitPoint.UpLimitType // 价格限制类型
	hasInit            bool                     // 是否已经预挂单初始化
}

func NewPre(accountKeyId uint8, symbolLen uint16, symbolIndex systemx.SymbolIndex16I, symbolKeyId uint64) *PointPre {
	return &PointPre{
		preAccountKeyId: accountKeyId,
		symbolLen:       symbolLen,
		symbolIndex:     symbolIndex,
		symbolKeyId:     symbolKeyId,
	}
}

func (s *PointPre) CheckPreOrder(symbolName string, markPrice float64, pScale systemx.PScale, qScale systemx.QScale) {
	if !s.hasInit {
		s.lastMarkPrice = markPrice
		if err := s.initPreOrder(qScale); err != nil {
			toUpBitDataStatic.DyLog.GetLog().Errorf("%s 初始化交易对失败:  %s", symbolName, err.Error())
			return
		}
		//下小订单
		if err := byBitOrderAppManager.GetTradeManager().SendPlaceOrder(s.preAccountKeyId, orderModel.MyPlaceOrderReq{
			SymbolName:    symbolName,
			ClientOrderId: s.clientOrderIdSmall,
			Pvalue:        u64Cal.FromF64(markPrice*toUpbitPoint.GetPointPre5(s.upT), pScale.Uint8()),
			Qvalue:        s.PointNum,
			Pscale:        pScale,
			Qscale:        qScale,
			OrderMode:     execute.SELL_OPEN_LIMIT,
			SymbolIndex:   s.symbolIndex,
			SymbolLen:     s.symbolLen,
			ReqFrom:       from_bybit,
			UsageFrom:     point_pre,
		}); err != nil {
			toUpBitDataStatic.DyLog.GetLog().Errorf("创建小订单错误: %s", err.Error())
		}
		return
	}
	// 价格距离上次变动有1%以上,开始移动订单
	diff := markPrice - s.lastMarkPrice
	if diff < 0 {
		diff = -diff
	}
	if diff >= 0.01*s.lastMarkPrice {
		lastMarkPrice := s.lastMarkPrice
		s.lastMarkPrice = markPrice
		smallPrice := u64Cal.FromF64(markPrice*toUpbitPoint.GetPointPre5(s.upT), pScale.Uint8())

		if _, ok := clientOrderOnline.Load(s.clientOrderIdSmall); ok {
			toUpBitDataStatic.DyLog.GetLog().Infof("[%d,%s] 触发[%.8f,%.8f,%d_%d],准备更新预挂单:%d", s.preAccountKeyId, symbolName, lastMarkPrice, markPrice, pScale, qScale, smallPrice)
			//更新订单价格
			if err := byBitOrderAppManager.GetTradeManager().SendModifyOrder(s.preAccountKeyId, orderModel.MyModifyOrderReq{
				SymbolName:    symbolName,
				ClientOrderId: s.clientOrderIdSmall,
				Pvalue:        smallPrice,
				Qvalue:        s.PointNum,
				Pscale:        pScale,
				Qscale:        qScale,
				OrderMode:     execute.SELL_OPEN_LIMIT,
				ReqFrom:       from_bybit,
				UsageFrom:     point_pre,
			}); err != nil {
				notifyTg.GetTg().SendToUpBitMsg(map[string]string{"symbol": symbolName, "op": "更新5%订单失败", "error": err.Error()})
				toUpBitDataStatic.DyLog.GetLog().Errorf("%s修改小订单错误: %s", symbolName, err.Error())
			}
		} else {
			toUpBitDataStatic.DyLog.GetLog().Infof("[%d,%s] 触发[%.8f,%.8f,%d_%d],准备重下预挂单:%d", s.preAccountKeyId, symbolName, s.lastMarkPrice, markPrice, pScale, qScale, smallPrice)
			if _, ok := ClientOrderNotOpen.Load(s.clientOrderIdSmall); ok {
				toUpBitDataStatic.DyLog.GetLog().Infof("订单限流中,无法下单:%s", string(s.clientOrderIdSmall[:]))
				return
			}
			//下小订单
			if err := byBitOrderAppManager.GetTradeManager().SendPlaceOrder(s.preAccountKeyId, orderModel.MyPlaceOrderReq{
				SymbolName:    symbolName,
				ClientOrderId: s.clientOrderIdSmall,
				Pvalue:        smallPrice,
				Qvalue:        s.PointNum,
				Pscale:        pScale,
				Qscale:        qScale,
				OrderMode:     execute.SELL_OPEN_LIMIT,
				SymbolIndex:   s.symbolIndex,
				SymbolLen:     s.symbolLen,
				ReqFrom:       from_bybit,
				UsageFrom:     point_pre,
			}); err != nil {
				notifyTg.GetTg().SendToUpBitMsg(map[string]string{"symbol": symbolName, "op": "预挂5%订单失败", "error": err.Error()})
				toUpBitDataStatic.DyLog.GetLog().Errorf("%s创建小订单错误: %s", symbolName, err.Error())
			}
		}
	}
}

func (s *PointPre) initPreOrder(qScale systemx.QScale) error {
	dyMeta, err := symbolDynamic.GetManager().Get(s.symbolKeyId)
	if err != nil {
		return err
	}
	//挂单量=最大(最小下单量,12*最小下单金额/最新标记价格)
	maxNum := math.Max(dyMeta.LotSize.InexactFloat64(), 2*dyMeta.MinQty.InexactFloat64()/(s.lastMarkPrice))
	s.PointNum = u64Cal.FromF64(maxNum, qScale.Uint8())
	s.clientOrderIdSmall = time2str.GetNowTimeStampMicroSlice16()
	s.hasInit = true
	return nil
}

func (s *PointPre) OnPreFilled(symbolName string, clientOrderId systemx.WsId16B, pScale systemx.PScale, qScale systemx.QScale) {
	safex.SafeGo("预挂单成交"+string(clientOrderId[:]), func() {
		if _, ok := clientOrderOnline.Load(clientOrderId); ok {
			// 是属于预挂单的订单,5秒后删除订单标记,防止频繁成交
			clientOrderOnline.Delete(clientOrderId)
			ClientOrderNotOpen.Store(clientOrderId, struct{}{})
			// 下买入平空单
			if err := byBitOrderAppManager.GetTradeManager().SendPlaceOrder(s.preAccountKeyId, orderModel.MyPlaceOrderReq{
				// 在1.03处下一个平空limit止盈单,只有挂单成功或者成交两种状态
				SymbolName:    symbolName,
				ClientOrderId: time2str.GetNowTimeStampMicroSlice16(),
				Pvalue:        u64Cal.FromF64(s.lastMarkPrice*toUpbitPoint.GetPointPre3(s.upT), pScale.Uint8()),
				Qvalue:        s.PointNum,
				Pscale:        pScale,
				Qscale:        qScale,
				OrderMode:     execute.BUY_CLOSE_LIMIT,
				SymbolIndex:   s.symbolIndex,
				SymbolLen:     s.symbolLen,
				ReqFrom:       from_bybit,
				UsageFrom:     point_pre,
			}); err != nil {
				notifyTg.GetTg().SendToUpBitMsg(map[string]string{"symbol": symbolName, "op": "下买入平空单失败"})
				toUpBitDataStatic.DyLog.GetLog().Errorf("%s下买入平空单错误: %s", symbolName, err.Error())
			}
			//刷新一下clientOrderId
			s.FreshClientOrderId()
			// 等待能再次下单
			time.Sleep(60 * time.Second)
			toUpBitDataStatic.DyLog.GetLog().Infof("预挂单成交60秒后,删除订单限流标记:%s", string(clientOrderId[:]))
			ClientOrderNotOpen.Delete(clientOrderId)
		}
	})
}

func (s *PointPre) CancelPreOrder(symbolName string, reqFrom instanceEnum.Type) {
	if !s.hasInit {
		return
	}
	byBitOrderAppManager.GetTradeManager().SendCancelOrder(s.preAccountKeyId, orderModel.MyQueryOrderReq{
		SymbolName:    symbolName,
		ClientOrderId: s.clientOrderIdSmall,
		ReqFrom:       reqFrom,
		UsageFrom:     point_pre,
	})
}

func (s *PointPre) FreshClientOrderId() {
	s.clientOrderIdSmall = time2str.GetNowTimeStampMicroSlice16()
}
