package orderModel

import (
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/instance/instanceDefine"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/server/usageEnum"
)

// 只改价格不改数量

type MyModifyOrderReq struct {
	SymbolName    string          //下单symbol
	ClientOrderId systemx.WsId16B //自己生产的id,交易计划的key
	Pvalue        systemx.OrderSdkType
	Qvalue        systemx.OrderSdkType
	Pscale        systemx.PScale
	Qscale        systemx.QScale
	OrderMode     execute.OrderMode   //订单模式BUY_OPEN,SELL_OPEN,BUY_CLOSE,SELL_CLOSE
	ReqFrom       instanceDefine.Type //实例枚举
	UsageFrom     usageEnum.Type      //用途枚举
}

func (s *MyModifyOrderReq) TypeName() string {
	return "MyModifyOrderReq"
}

// Check 下单非空判断
func (s *MyModifyOrderReq) Check() error {
	return nil
}
