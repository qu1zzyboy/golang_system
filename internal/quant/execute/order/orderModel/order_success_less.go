package orderModel

import (
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/wsRequestCache"
	"upbitBnServer/server/usageEnum"
)

type OnSuccessEvt struct {
	ClientOrderId systemx.WsId16B //必要的
	Price         float64
	Volume        float64
	T             int64 //必要的
	OrderMode     execute.OrderMode
	AccountKeyId  uint8
	UsageFrom     usageEnum.Type      //实例的用途枚举
	OrderStatus   execute.OrderStatus //必要的
}

type OnFailedEvt struct {
	ClientOrderId systemx.WsId16B              // 客户端订单ID
	P             float64                      // 4016探针返回价格上下界限
	ErrorCode     int32                        // 错误码
	UsageFrom     usageEnum.Type               //用途枚举
	ReqType       wsRequestCache.WsRequestType //请求接口类型
	AccountKeyId  uint8
}
