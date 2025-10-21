package orderModel

import (
	"github.com/hhh500/quantGoInfra/infra/errorx/errDefine"
	"github.com/hhh500/quantGoInfra/resource/resourceEnum"
	"github.com/hhh500/upbitBnServer/internal/quant/execute"
	"github.com/hhh500/upbitBnServer/internal/quant/market/symbolInfo/symbolStatic"
	"github.com/shopspring/decimal"
)

// 只改价格不改数量

type MyModifyOrderReq struct {
	DepthOnePrice decimal.Decimal           //深度1价格
	OrigPrice     decimal.Decimal           //原订单价格
	OrigVol       decimal.Decimal           //订单数量
	ModifyPrice   decimal.Decimal           //订单价格
	ClientOrderId string                    //自己生产的id,交易计划的key
	StaticMeta    *symbolStatic.StaticTrade //交易对静态数据
	OrderMode     execute.MyOrderMode       //订单模式BUY_OPEN,SELL_OPEN,BUY_CLOSE,SELL_CLOSE
	From          resourceEnum.ResourceFrom // 订单来源
}

func (s *MyModifyOrderReq) TypeName() string {
	return "MyModifyOrderReq"
}

// Check 下单非空判断
func (s *MyModifyOrderReq) Check() error {
	if s.ClientOrderId == "" {
		return errDefine.ClientOrderIdEmpty
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
	if s.ModifyPrice.LessThanOrEqual(decimal.Zero) {
		return errDefine.ValueInvalid.WithMetadata(map[string]string{modifyPrice: s.ModifyPrice.String()})
	}
	return nil
}

type MyModifyOrderBatchReq struct {
	StaticMeta *symbolStatic.StaticTrade
	Orders     []*MyModifyOrderReq
}
