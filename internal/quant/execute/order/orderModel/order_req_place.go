package orderModel

import (
	"github.com/hhh500/quantGoInfra/infra/errorx/errDefine"
	"github.com/hhh500/quantGoInfra/resource/resourceEnum"
	"github.com/hhh500/upbitBnServer/internal/quant/execute"
	"github.com/hhh500/upbitBnServer/internal/quant/market/symbolInfo/symbolStatic"
	"github.com/shopspring/decimal"
)

const (
	origPrice   = "origPrice"   // 原始价格
	origVol     = "origVol"     // 原始数量
	modifyPrice = "modifyPrice" // 修改后的价格
)

type MyPlaceOrderReq struct {
	OrigPrice     decimal.Decimal           //订单价格
	OrigVol       decimal.Decimal           //订单数量
	ClientOrderId string                    //自己生产的id,交易计划的key
	StaticMeta    *symbolStatic.StaticTrade //静态交易对信息
	OrderType     execute.MyOrderPlaceType  //下单类型 LIMIT,ioc,post_only,自定义字段
	OrderMode     execute.MyOrderMode       //订单模式BUY_OPEN...下单打印会用到,自动赋值
	From          resourceEnum.ResourceFrom // 订单来源
}

func (s *MyPlaceOrderReq) TypeName() string {
	return "MyPlaceOrderReq"
}

// Check 下单非空判断
func (s *MyPlaceOrderReq) Check() error {
	if s.ClientOrderId == "" {
		return errDefine.ClientOrderIdEmpty
	}
	if err := s.OrderType.Verify(); err != nil {
		return err
	}
	if err := s.OrderMode.Verify(); err != nil {
		return err
	}
	if s.OrigPrice.LessThanOrEqual(decimal.Zero) {
		return errDefine.ValueInvalid.WithMetadata(map[string]string{origPrice: s.OrigPrice.String()})
	}
	if s.OrigVol.LessThanOrEqual(decimal.Zero) {
		return errDefine.ValueInvalid.WithMetadata(map[string]string{origVol: s.OrigVol.String()})
	}
	return nil
}

// SendErrorMsg 订单异常发生提醒
func (s *MyPlaceOrderReq) SendErrorMsg(errCode int64, errReason string) {
	// // 限流,防止同一个类型的错误一直发送
	// if !rateLimitUtils.ErrorMsgLimit.Allow(errCode) {
	// 	return
	// }
	// r := fmt.Sprintf("下单错误,来源:%s_%s\n异常时间:%s\n错误实例:%s\n交易key: %s\n下单价: %s\n下单量: %s\n挂单方式:%s\n错误码:%d,错误原因: %s",
	// 	conf.ServerName, "GOLANG", timeUtils.GetNowTimeStr(), "", s.StaticMeta.GetSymbolKey(),
	// 	s.OrigPrice.String(), s.OrigVol.String(),
	// 	s.OrderType.String(), errCode, errReason)
	// msgUtils.SendNormalErrorMsg("下单错误", r)
}

type MyPlaceOrderReqSortLtSlice []*MyPlaceOrderReq

func (tp MyPlaceOrderReqSortLtSlice) Len() int {
	return len(tp)
}

func (tp MyPlaceOrderReqSortLtSlice) Swap(i, j int) {
	tp[i], tp[j] = tp[j], tp[i]
}

func (tp MyPlaceOrderReqSortLtSlice) Less(i, j int) bool {
	return tp[i].OrigPrice.InexactFloat64() < tp[j].OrigPrice.InexactFloat64()
}

type MyPlaceOrderReqSortGtSlice []*MyPlaceOrderReq

func (tp MyPlaceOrderReqSortGtSlice) Len() int {
	return len(tp)
}

func (tp MyPlaceOrderReqSortGtSlice) Swap(i, j int) {
	tp[i], tp[j] = tp[j], tp[i]
}

func (tp MyPlaceOrderReqSortGtSlice) Less(i, j int) bool {
	return tp[i].OrigPrice.InexactFloat64() > tp[j].OrigPrice.InexactFloat64()
}
