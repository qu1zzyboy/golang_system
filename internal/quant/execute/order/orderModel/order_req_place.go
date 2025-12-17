package orderModel

import (
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/instance/instanceDefine"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/server/usageEnum"
)

const (
	origPrice   = "origPrice"   // 原始价格
	origVol     = "origVol"     // 原始数量
	modifyPrice = "modifyPrice" // 修改后的价格
)

type MyPlaceOrderReq struct {
	SymbolName    string                 //下单symbol
	ClientOrderId systemx.WsId16B        //自己生产的id,交易计划的key
	Pvalue        systemx.OrderSdkType   //定点价格
	Qvalue        systemx.OrderSdkType   //定点数量
	SymbolIndex   systemx.SymbolIndex16I //交易对的唯一标识
	Pscale        systemx.PScale         //
	Qscale        systemx.QScale         //
	OrderMode     execute.OrderMode      //订单模式BUY_OPEN_LIMIT...
	ReqFrom       instanceDefine.Type    //实例枚举
	UsageFrom     usageEnum.Type         //用途枚举
}

func (s *MyPlaceOrderReq) TypeName() string {
	return "MyPlaceOrderReq"
}

// Check 下单非空判断
func (s *MyPlaceOrderReq) Check() error {
	return nil
}
