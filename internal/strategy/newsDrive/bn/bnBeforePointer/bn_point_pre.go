package bnBeforePointer

import (
	"math"
	"sync/atomic"
	"time"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/quant/market/symbolInfo"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolDynamic"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolStatic"
	"upbitBnServer/internal/strategy/newsDrive/bn/bnDriveOrderWrap"
	"upbitBnServer/internal/strategy/newsDrive/common/driverDefine"
	"upbitBnServer/internal/strategy/newsDrive/common/driverStatic"
	"upbitBnServer/pkg/utils/time2str"
	"upbitBnServer/server/instanceEnum"
	"upbitBnServer/server/usageEnum"

	"github.com/shopspring/decimal"
)

const (
	point_pre = usageEnum.NEWS_DRIVE_PRE
	from_bn   = instanceEnum.DRIVER_LIST_BN
)

/*
目前一共只会被两个线程调用

主线程: 接收markPrice变动，更新探针订单价格、设置探针挂单状态
成交平仓线程:成交之后设置限流标志、下平仓单

SELL_OPEN_LIMIT ---> NEWS_DRIVE_PRE
BUY_CLOSE_LIMIT ---> POINTER_BIDS_3
*/

type PointPreBN struct {
	symbolName         string                   // 交易对名称
	clientOrderIdSmall systemx.WsId16B          // 探针id
	lastMarkPrice      float64                  // 上次标记价格,1倍
	PointNum           systemx.OrderSdkType     // 预挂单订单数量,放大了10的qScale次方
	isOrderLimit       atomic.Bool              // 订单是否在限流中,多线程访问
	pScale             systemx.PScale           // 价格小数位
	qScale             systemx.QScale           // 数量小数位
	symbolLen          uint16                   // symbol长度
	symbolIndex        systemx.SymbolIndex16I   // 交易对下标
	accountKeyId       uint8                    // 预挂单的账户id
	upT                driverDefine.UpLimitType // 价格限制类型
	hasInit            bool                     // 是否已经预挂单初始化
	isPointOnline      bool                     // 探针订单是否在挂单状态,只由主线程调用
}

func NewPreBN(pScale systemx.PScale, qScale systemx.QScale, accountKeyId uint8, symbolLen uint16, symbolIndex systemx.SymbolIndex16I, symbolName string) *PointPreBN {
	return &PointPreBN{
		symbolName:   symbolName,
		symbolLen:    symbolLen,
		symbolIndex:  symbolIndex,
		pScale:       pScale,
		qScale:       qScale,
		accountKeyId: accountKeyId,
	}
}

func (s *PointPreBN) GetPointNum() systemx.OrderSdkType {
	return s.PointNum
}

func (s *PointPreBN) CancelPreOrder() {
	bnDriveOrderWrap.CancelOrderWithPlan(s.accountKeyId, &orderModel.MyQueryOrderReq{
		SymbolName:    s.symbolName,
		ClientOrderId: s.clientOrderIdSmall,
		ReqFrom:       from_bn,
		UsageFrom:     point_pre,
	})
}

// ReceiveSuOrder 主线程接收订单返回调用,更新订单状态
func (s *PointPreBN) ReceiveSuOrder(isOnline bool, clientOrderId systemx.WsId16B) {
	if clientOrderId == s.clientOrderIdSmall {
		s.isPointOnline = isOnline
	}
}

func (s *PointPreBN) CheckPreOrder(markPrice float64) {
	if !s.hasInit {
		s.lastMarkPrice = markPrice
		if err := s.initPreOrder(); err != nil {
			driverStatic.DyLog.GetLog().Errorf("%s 初始化交易对失败:  %s", s.symbolName, err.Error())
			return
		}
		//下小订单
		bnDriveOrderWrap.PlaceOrderWithPlan(s.accountKeyId, &orderModel.MyPlaceOrderReq{
			SymbolName:    s.symbolName,
			ClientOrderId: s.clientOrderIdSmall,
			Pvalue:        systemx.OrderSdkType(decimal.NewFromFloat(markPrice * driverDefine.GetPointPre5(s.upT)).Truncate(s.pScale.Uint8())),
			Qvalue:        s.PointNum,
			Pscale:        s.pScale,
			Qscale:        s.qScale,
			OrderMode:     execute.SELL_OPEN_LIMIT,
			SymbolIndex:   s.symbolIndex,
			ReqFrom:       from_bn,
			UsageFrom:     point_pre,
		})
		return
	}
	// 价格距离上次变动有1%以上,开始移动订单
	diff := markPrice - s.lastMarkPrice
	if diff < 0 {
		diff = -diff
	}
	if diff < 0.01*s.lastMarkPrice {
		return
	}

	lastMarkPrice := s.lastMarkPrice
	s.lastMarkPrice = markPrice
	smallPrice := systemx.OrderSdkType(decimal.NewFromFloat(markPrice * driverDefine.GetPointPre5(s.upT)).Truncate(s.pScale.Uint8()))

	if s.isPointOnline {
		driverStatic.DyLog.GetLog().Infof("[%d,%s] 触发[%.8f,%.8f,%d_%d],准备更新预挂单:%d", s.accountKeyId, s.symbolName, lastMarkPrice, markPrice, s.pScale, s.qScale, smallPrice)
		//更新订单价格
		bnDriveOrderWrap.ModifyOrderWithPlan(s.accountKeyId, &orderModel.MyModifyOrderReq{
			SymbolName:    s.symbolName,
			ClientOrderId: s.clientOrderIdSmall,
			Pvalue:        smallPrice,
			Qvalue:        s.PointNum,
			Pscale:        s.pScale,
			Qscale:        s.qScale,
			OrderMode:     execute.SELL_OPEN_LIMIT,
			ReqFrom:       from_bn,
			UsageFrom:     point_pre,
		})
	} else {
		driverStatic.DyLog.GetLog().Infof("[%d,%s] 触发[%.8f,%.8f,%d_%d],准备重下预挂单:%d", s.accountKeyId, s.symbolName, s.lastMarkPrice, markPrice, s.pScale, s.qScale, smallPrice)
		if s.isOrderLimit.Load() {
			driverStatic.DyLog.GetLog().Infof("订单限流中,无法下单:%s", string(s.clientOrderIdSmall[:]))
			return
		}
		//下小订单
		bnDriveOrderWrap.PlaceOrderWithPlan(s.accountKeyId, &orderModel.MyPlaceOrderReq{
			SymbolName:    s.symbolName,
			ClientOrderId: s.clientOrderIdSmall,
			Pvalue:        smallPrice,
			Qvalue:        s.PointNum,
			Pscale:        s.pScale,
			Qscale:        s.qScale,
			OrderMode:     execute.SELL_OPEN_LIMIT,
			SymbolIndex:   s.symbolIndex,
			ReqFrom:       from_bn,
			UsageFrom:     point_pre,
		})
	}
}

func (s *PointPreBN) initPreOrder() error {
	symbolId, ok := symbolStatic.GetSymbol().GetSymbol(s.symbolName)
	if !ok {
		return exchangeEnum.BINANCE.GetNotSupportError("symbolId not found")
	}
	symbolKeyId := symbolInfo.MakeSymbolKey3(exchangeEnum.BINANCE, exchangeEnum.FUTURE, symbolId)
	dyMeta, err := symbolDynamic.GetManager().Get(symbolKeyId)
	if err != nil {
		return err
	}
	//挂单量=最大(最小下单量,12*最小下单金额/最新标记价格)
	maxNum := math.Max(dyMeta.LotSize.InexactFloat64(), 12*dyMeta.MinQty.InexactFloat64()/(s.lastMarkPrice))
	s.PointNum = systemx.OrderSdkType(decimal.NewFromFloat(maxNum).Truncate(s.qScale.Uint8()))
	s.clientOrderIdSmall = time2str.GetNowTimeStampMicroSlice16()
	s.hasInit = true
	return nil
}

func (s *PointPreBN) OnPreFilled(clientOrderId systemx.WsId16B) {
	safex.SafeGo("预挂单成交"+string(clientOrderId[:]), func() {
		if clientOrderId != s.clientOrderIdSmall {
			driverStatic.DyLog.GetLog().Infof("预挂单成交,但clientOrderId不匹配:%s", string(clientOrderId[:]))
			return
		}
		s.isOrderLimit.Store(true)
		// 下买入平空单
		if err := bnDriveOrderWrap.PlaceOrderWithPlan(s.accountKeyId, &orderModel.MyPlaceOrderReq{
			// 在1.03处下一个平空limit止盈单,只有挂单成功或者成交两种状态
			SymbolName:    s.symbolName,
			ClientOrderId: time2str.GetNowTimeStampMicroSlice16(),
			Pvalue:        systemx.OrderSdkType(decimal.NewFromFloat(s.lastMarkPrice * driverDefine.GetPointPre3(s.upT)).Truncate(s.pScale.Uint8())),
			Qvalue:        s.PointNum,
			Pscale:        s.pScale,
			Qscale:        s.qScale,
			OrderMode:     execute.BUY_CLOSE_LIMIT,
			SymbolIndex:   s.symbolIndex,
			ReqFrom:       from_bn,
			UsageFrom:     point_pre,
		}); err != nil {
			driverStatic.SendToUpBitMsg("下买入平空单错误", map[string]string{"symbol": s.symbolName, "op": "下买入平空单失败"})
		}
		//刷新一下clientOrderId
		s.clientOrderIdSmall = time2str.GetNowTimeStampMicroSlice16()
		// 等待能再次下单
		time.Sleep(60 * time.Second)
		s.isOrderLimit.Store(false)
		driverStatic.DyLog.GetLog().Infof("预挂单成交60秒后,删除订单限流标记:%s", string(clientOrderId[:]))
	})
}
