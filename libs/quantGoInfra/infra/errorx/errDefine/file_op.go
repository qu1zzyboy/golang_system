package errDefine

import (
	"github.com/hhh500/quantGoInfra/infra/errorx"
	"github.com/hhh500/quantGoInfra/infra/errorx/errCode"
)

var (
	DelRemoteFileErr = errorx.New(errCode.DEL_REMOTE_FILE_ERR, "删除远程文件失败")
)
