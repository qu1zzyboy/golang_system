package staticLog

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"

	"upbitBnServer/internal/define/defineTime"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"gopkg.in/natefinch/lumberjack.v2"
)

func GetLogFileName(fileDir, dateStr, fileName string) string {
	p := "./logs"

	if dateStr != "" {
		p = filepath.Join(p, dateStr)
	}

	if fileDir != "" {
		p = filepath.Join(p, fileDir)
	}

	return filepath.Join(p, fmt.Sprintf("%s.log", fileName))
}

func setLogDefaultConfig(log *logrus.Logger) {
	log.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: defineTime.FormatMillSec, //时间格式
		FullTimestamp:   true,                     //完整时间戳
		ForceColors:     true,                     //强制使用颜色
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			fileName := "[" + path.Base(frame.File) + ":" + strconv.Itoa(frame.Line) + "]"
			return "", fileName
		},
	})
	log.SetReportCaller(true) //开启文件名和行号
}

func NewLoggerWithLever(cfg Config) *logrus.Logger {
	log := logrus.New()
	log.SetOutput(io.Discard) //关闭默认输出 采用钩子方式输出到文件及控制台

	// DEBUG 级别只输出到控制台，不写入文件
	log.AddHook(&writer.Hook{
		Writer:    os.Stdout,
		LogLevels: []logrus.Level{logrus.DebugLevel},
	})

	// INFO 及以上级别输出到控制台和文件
	lumberjackOutputInfo := &lumberjack.Logger{
		Filename:   GetLogFileName(cfg.FileDir, cfg.DateStr, cfg.FileName),
		MaxSize:    MaxFileSize,
		MaxBackups: MaxBackups,
		MaxAge:     MaxAliveAge,
		Compress:   IsCompress,
	}
	log.AddHook(&writer.Hook{
		Writer:    io.MultiWriter(os.Stdout, lumberjackOutputInfo),
		LogLevels: []logrus.Level{logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel},
	})
	//添加error钩子
	if cfg.NeedErrorHook {
		lumberjackOutputError := &lumberjack.Logger{
			Filename:   GetLogFileName(ErrorDir, cfg.DateStr, ErrorSum),
			MaxSize:    MaxFileSize,
			MaxBackups: MaxBackups,
			MaxAge:     MaxAliveAge,
			Compress:   IsCompress,
		}
		log.AddHook(&writer.Hook{
			Writer:    io.MultiWriter(lumberjackOutputError),
			LogLevels: []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel},
		})
	}
	log.SetLevel(GetLogLevelFromEnum(cfg.Level)) //设置日志级别
	setLogDefaultConfig(log)
	return log
}
