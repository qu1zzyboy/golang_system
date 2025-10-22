package errDefine

import (
	"github.com/hhh500/quantGoInfra/infra/errorx"
	"github.com/hhh500/quantGoInfra/infra/errorx/errCode"
)

var (
	PointerNil      = errorx.New(errCode.POINTER_NIL, "空指针错误")
	EnumDefineError = errorx.New(errCode.ENUM_DEFINE_ERROR, "枚举未定义错误")
	ValueInvalid    = errorx.New(errCode.INVALID_VALUE, "值异常")
)
