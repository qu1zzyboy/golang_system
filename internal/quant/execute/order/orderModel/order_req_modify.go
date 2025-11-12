package orderModel

import (
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/instanceEnum"
	"upbitBnServer/internal/infra/systemx/usageEnum"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolStatic"
)

// 只改价格不改数量

type MyModifyOrderReq struct {
	SymbolName    string          //下单symbol
	ClientOrderId systemx.WsId16B //自己生产的id,交易计划的key
	Pvalue        uint64
	Qvalue        uint64
	Pscale        systemx.PScale
	Qscale        systemx.QScale
	OrderMode     execute.OrderMode //订单模式BUY_OPEN,SELL_OPEN,BUY_CLOSE,SELL_CLOSE
	ReqFrom       instanceEnum.Type //实例枚举
	UsageFrom     usageEnum.Type    //用途枚举
}

func (s *MyModifyOrderReq) TypeName() string {
	return "MyModifyOrderReq"
}

type MyModifyOrderBatchReq struct {
	StaticMeta *symbolStatic.StaticTrade
	Orders     []*MyModifyOrderReq
}
