package instanceDefine

import "context"

type Type uint8

const (
	SEND_SUCCESS_ORDER    Type = 127 //成功订单
	SYMBOL_ON_LIST        Type = 126 //交易对上线
	SYMBOL_DOWN_LIST      Type = 125 //交易对下线
	SYMBOL_DYNAMIC_CHANGE Type = 124 //交易对动态变动
)

type Update struct {
	JsonData string `json:"jsonData"`
}

type Instance interface {
	OnStop(ctx context.Context) error
	OnUpdate(ctx context.Context, param Update) error
}

// MakeInstanceId 构造唯一的实例id
func MakeInstanceId(symbolIndex uint16, instanceType uint8, accountKeyId uint8) uint32 {
	return (uint32(symbolIndex) << 16) | (uint32(instanceType) << 8) | uint32(accountKeyId)
}
