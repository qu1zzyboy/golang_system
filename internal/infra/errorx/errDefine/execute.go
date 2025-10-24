package errDefine

import (
	"upbitBnServer/internal/infra/errorx"
	"upbitBnServer/internal/infra/errorx/errCode"
)

var (
	InstanceEmpty          = errorx.New(errCode.INSTANCE_ID_EMPTY, "instanceId 不能为空")
	InstanceNotExists      = errorx.New(errCode.INSTANCE_NOT_EXISTS, "instanceId 不存在")
	InstanceExists         = errorx.New(errCode.INSTANCE_EXISTS, "instanceId 已存在")
	ClientOrderIdEmpty     = errorx.New(errCode.CLIENT_ORDER_ID_EMPTY, "clientOrderId 不能为空")
	ClientOrderIdExists    = errorx.New(errCode.CLIENT_ORDER_ID_EXISTS, "clientOrderId 已存在")
	ClientOrderIdNotExists = errorx.New(errCode.CLIENT_ORDER_ID_NOT_EXISTS, "clientOrderId 不存在")

	AccountKeyEmpty     = errorx.New(errCode.ACCOUNT_KEY_EMPTY, "accountKey 不能为空")
	AccountKeyNotExists = errorx.New(errCode.ACCOUNT_KEY_NOT_EXISTS, "accountKey 不存在")
	SymbolKeyEmpty      = errorx.New(errCode.SYMBOL_KEY_EMPTY, "symbolKey 不能为空")
)

var (
	OrderStatusEmpty = errorx.New(errCode.ORDER_STATUS_EMPTY, "orderStatus 不能为空")
	OrderFromEmpty   = errorx.New(errCode.FROM_EMPTY, "From 不能为空")
)
