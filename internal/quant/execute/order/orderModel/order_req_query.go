package orderModel

import (
	"github.com/hhh500/quantGoInfra/infra/errorx/errDefine"
	"github.com/hhh500/quantGoInfra/resource/resourceEnum"
	"github.com/hhh500/upbitBnServer/internal/quant/market/symbolInfo/symbolStatic"
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
