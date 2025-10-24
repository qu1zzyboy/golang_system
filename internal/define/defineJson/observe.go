package defineJson

const (
	ErrCode   = "errCode"   // 错误码
	ErrReason = "errReason" // 错误原因
	ErrMsg    = "errMsg"    // 错误信息
	ErrStack  = "errStack"  // 错误堆栈信息
)

const (
	LogFileName = "logFile"   // 日志文件
	LogPath     = "logPath"   // 日志路径
	Msg         = "msg"       // 通用消息
	Level       = "level"     // 日志级别
	Time        = "time"      // 日志时间
	Caller      = "caller"    // 日志调用者
	TimeStamp   = "timeStamp" // 日志时间戳
)

const (
	QuantSystem = "quant_system"
	TraceId     = "trace_id"   // 跟踪id,用于追踪请求链路,每个请求一个
	ProtectId   = "protect_id" // 保护id,用于标识保护任务,每个任务一个
	UserId      = "user_id"    // 用户id
	TaskId      = "task_id"    // 任务标识，用于区分不同的异步任务类型
)
