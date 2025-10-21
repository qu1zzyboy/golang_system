package orderModel

import (
	"context"
)

type FailedOrder struct {
	From          string //来源,数据产生的地方赋值
	ClientOrderId string //clientOrderId订单标识
	ErrMsg        string //错误信息
	ErrReason     string //错误码
	AccountKeyId  uint8  //自添加:用户key
}

type FailedOrderTrace struct {
	Ctx context.Context //上下文
	Su  *FailedOrder    //订单对象
}

func NewFailedOrder(clientOrderId, reason, msg string) *FailedOrder {
	return &FailedOrder{
		ClientOrderId: clientOrderId,
		ErrMsg:        msg,
		ErrReason:     reason,
	}
}

func NewFailedOrderTrace(ctx context.Context, su *FailedOrder) *FailedOrderTrace {
	return &FailedOrderTrace{
		Ctx: ctx,
		Su:  su,
	}
}

type FailureOrderHandler interface {
	OnFailureOrder(ctx context.Context, fa *FailedOrderTrace) error
}
