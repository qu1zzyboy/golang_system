package errDefine

import (
	"upbitBnServer/internal/infra/errorx"
	"upbitBnServer/internal/infra/errorx/errCode"
)

var (
	DelRemoteFileErr = errorx.New(errCode.DEL_REMOTE_FILE_ERR, "删除远程文件失败")
)
