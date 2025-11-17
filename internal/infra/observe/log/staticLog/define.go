package staticLog

import (
	"upbitBnServer/internal/infra/observe/log/logCfg"
)

type Config struct {
	NeedErrorHook bool            // 是否需要错误日志钩子
	FileDir       string          // 日志文件夹
	DateStr       string          // 日志日期字符串
	FileName      string          // 日志文件名
	Level         logCfg.LogLevel // 日志级别
}

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
		Level:         logCfg.G_LOG_LEVEL,
	}
	Log      = NewLoggerWithLever(defaultCfg)
	LogPanic = NewLoggerWithLever(Config{
		NeedErrorHook: true,
		FileDir:       "",
		DateStr:       "",
		FileName:      "panic",
		Level:         logCfg.INFO_LEVEL,
	})
)
