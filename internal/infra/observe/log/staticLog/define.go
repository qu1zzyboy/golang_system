package staticLog

import (
	"github.com/sirupsen/logrus"
)

type Config struct {
	NeedErrorHook bool     // 是否需要错误日志钩子
	FileDir       string   // 日志文件夹
	DateStr       string   // 日志日期字符串
	FileName      string   // 日志文件名
	Level         LogLevel // 日志级别
}

type LogLevel uint8

const (
	DEBUG_LEVEL LogLevel = iota //用于开发和调试阶段的详细信息
	INFO_LEVEL                  //用于记录程序的常规操作和状态更新
	WARN_LEVEL                  //表示某些非预期的事件或潜在问题,但系统仍能继续正常运行。
	ERROR_LEVEL                 //表示发生了错误,影响了系统的某个部分,导致某些功能无法正常运行。
	FATAL_LEVEL                 //表示严重错误,通常意味着程序的执行已经无法继续,必须立即终止。
	PANIC_LEVEL                 //表示比 FATAL 更严重的错误。通常会导致程序异常终止（panic）
)

const (
	MaxFileSize = 500         // 文件大小,单位MB
	MaxBackups  = 3           // 最多保留3个备份
	MaxAliveAge = 28          // 最大保留28天
	IsCompress  = false       // 是否压缩
	ErrorSum    = "error_sum" // 错误日志文件名
	ErrorDir    = "error"     // 错误日志文件夹名
)

var (
	defaultCfg = Config{
		NeedErrorHook: true,
		FileDir:       "",
		DateStr:       "",
		FileName:      "normal",
		Level:         DEBUG_LEVEL,
	}
	Log      = NewLoggerWithLever(defaultCfg)
	LogPanic = NewLoggerWithLever(Config{
		NeedErrorHook: true,
		FileDir:       "",
		DateStr:       "",
		FileName:      "panic",
		Level:         INFO_LEVEL,
	})
)

func GetLogLevelFromEnum(logLevel LogLevel) logrus.Level {
	switch logLevel {
	case PANIC_LEVEL:
		return logrus.PanicLevel
	case FATAL_LEVEL:
		return logrus.FatalLevel
	case ERROR_LEVEL:
		return logrus.ErrorLevel
	case WARN_LEVEL:
		return logrus.WarnLevel
	case INFO_LEVEL:
		return logrus.InfoLevel
	case DEBUG_LEVEL:
		return logrus.DebugLevel
	default:
		return logrus.DebugLevel
	}
}
