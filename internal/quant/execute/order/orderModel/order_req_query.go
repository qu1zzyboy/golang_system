package orderModel

import (
	"upbitBnServer/internal/infra/errorx/errDefine"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolStatic"
	"upbitBnServer/internal/resource/resourceEnum"
)

type MyQueryOrderReq struct {
	StaticMeta    *symbolStatic.StaticTrade //交易对静态数据
	ClientOrderId string                    //自己生产的id,交易计划的key
	From          resourceEnum.ResourceFrom // 订单来源
}

func (s *MyQueryOrderReq) TypeName() string {
	return "MyQueryOrderReq"
}

func (s *MyQueryOrderReq) Check() error {
	if s.ClientOrderId == "" {
		return errDefine.ClientOrderIdEmpty
	}
	return nil
}
