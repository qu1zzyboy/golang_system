package orderModel

import (
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/instanceEnum"
	"upbitBnServer/internal/infra/systemx/usageEnum"
	"upbitBnServer/internal/quant/execute"
)

const (
	origPrice   = "origPrice"   // 原始价格
	origVol     = "origVol"     // 原始数量
	modifyPrice = "modifyPrice" // 修改后的价格
)

type MyPlaceOrderReq struct {
	SymbolName    string                 //下单symbol
	ClientOrderId systemx.WsId16B        //自己生产的id,交易计划的key
	Pvalue        uint64                 //定点价格
	Qvalue        uint64                 //定点数量
	Pscale        systemx.PScale         //
	Qscale        systemx.QScale         //
	OrderMode     execute.OrderMode      //订单模式BUY_OPEN_LIMIT...
	SymbolIndex   systemx.SymbolIndex16I //交易对的唯一标识
	SymbolLen     uint16                 //交易对长度
	ReqFrom       instanceEnum.Type      //实例枚举
	UsageFrom     usageEnum.Type         //用途枚举
}

func (s *MyPlaceOrderReq) TypeName() string {
	return "MyPlaceOrderReq"
}

// Check 下单非空判断
func (s *MyPlaceOrderReq) Check() error {
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
	return tp[i].Pvalue < tp[j].Pvalue
}

type MyPlaceOrderReqSortGtSlice []*MyPlaceOrderReq

func (tp MyPlaceOrderReqSortGtSlice) Len() int {
	return len(tp)
}

func (tp MyPlaceOrderReqSortGtSlice) Swap(i, j int) {
	tp[i], tp[j] = tp[j], tp[i]
}

func (tp MyPlaceOrderReqSortGtSlice) Less(i, j int) bool {
	return tp[i].Pvalue > tp[j].Pvalue
}
