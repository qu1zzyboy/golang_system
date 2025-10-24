package errConst

const (
	ReasonFileOpenErr = "FILE_OPEN_ERR"     // 文件打开错误,如日志文件打开失败等
	MsgFileOpenErr    = "文件打开失败，请检查文件路径和权限" // 文件打开错误提示信息,如日志文件打开失败等

	MsgDirOpenErr = "目录打开失败，请检查目录路径和权限"

	ReasonLogNotFound = "LOG_NOT_FOUND" // 日志未找到错误,如日志文件不存在等
	MsgLogNotFound    = "日志对象未找到,请先初始化日志对象"

	ReasonLogNewFailed = "LOG_NEW_FAILED"      // 日志对象创建失败,如日志初始化失败等
	MsgLogNewFailed    = "日志对象创建失败,请检查日志配置和权限" // 日志对象创建失败提示信息,如日志初始化失败等
)
