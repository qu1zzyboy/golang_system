package wsSub

import (
	"github.com/hhh500/quantGoInfra/infra/errorx"
	"github.com/hhh500/quantGoInfra/infra/errorx/errCode"
)

const (
	sub_ADD  = "add_sub"
	sub_DEL  = "del_sub"
	sub_LIST = "list_sub"
	sub_AUTH = "auth_sub"
)

var (
	connErr = errorx.Newf(errCode.CodeWsDoError, "WS_CONN_ERROR", "ws行情数据连接失败")
)
