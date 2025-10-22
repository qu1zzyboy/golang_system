package errDefine

import (
	"github.com/hhh500/quantGoInfra/infra/errorx"
	"github.com/hhh500/quantGoInfra/infra/errorx/errCode"
)

var (
	HttpDoError     = errorx.New(errCode.HTTP_DO_ERROR, "HTTP请求错误")
	HttpParamError  = errorx.New(errCode.HTTP_PARAM_ERROR, "HTTP参数错误")
	JsonNotExpected = errorx.New(errCode.JSON_NOT_EXPECTED, "不被预期的json")
)
