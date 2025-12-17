package orderModel

import (
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/instance/instanceDefine"
	"upbitBnServer/server/usageEnum"
)

type MyQueryOrderReq struct {
	SymbolName    string              //下单symbol
	ClientOrderId systemx.WsId16B     //自己生产的id,交易计划的key
	ReqFrom       instanceDefine.Type //实例枚举
	UsageFrom     usageEnum.Type      //用途枚举
}

func (s *MyQueryOrderReq) TypeName() string {
	return "MyQueryOrderReq"
}

func (s *MyQueryOrderReq) Check() error {
	return nil
}
