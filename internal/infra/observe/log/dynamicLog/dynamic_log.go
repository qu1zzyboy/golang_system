package dynamicLog

import (
	"fmt"
	"sync/atomic"

	"upbitBnServer/pkg/container/map/myMap"

	"upbitBnServer/pkg/utils/timeUtils"

	"upbitBnServer/internal/infra/observe/log/staticLog"

	"github.com/sirupsen/logrus"
)

var (
	allLogMap = myMap.NewMySyncMap[string, *DynamicLogger]()
	Log       = NewDynamicLogger(staticLog.Config{
		NeedErrorHook: true,
		FileDir:       "",
		DateStr:       timeUtils.GetNowDateStr(),
		FileName:      "normal",
		Level:         staticLog.DEBUG_LEVEL,
	})
	Error = NewDynamicLogger(staticLog.Config{
		NeedErrorHook: false,
		FileDir:       "",
		DateStr:       timeUtils.GetNowDateStr(),
		FileName:      staticLog.ErrorSum,
		Level:         staticLog.ERROR_LEVEL,
	})
	Latency = NewDynamicLogger(staticLog.Config{
		NeedErrorHook: false,
		FileDir:       "latency",
		DateStr:       timeUtils.GetNowDateStr(),
		FileName:      "",
		Level:         staticLog.ERROR_LEVEL,
	})
)

type DynamicLogger struct {
	cfg staticLog.Config // 日志配置
	log atomic.Value     // 存储 *logrus.Logger
}

func NewDynamicLogger(cfg staticLog.Config) *DynamicLogger {
	s := DynamicLogger{
		cfg: cfg,
		log: atomic.Value{},
	}
	s.log.Store(staticLog.NewLoggerWithLever(cfg))
	allLogMap.Store(fmt.Sprintf("%s_%s", cfg.FileDir, cfg.FileName), &s)
	return &s
}

func (s *DynamicLogger) GetLog() *logrus.Logger {
	return s.log.Load().(*logrus.Logger)
}

func (s *DynamicLogger) RefreshLogConfig(dataStr string) {
	s.log.Store(staticLog.NewLoggerWithLever(s.cfg))
}

func refreshAllLog(dataStr string) {
	staticLog.Log.Info("遍历所有日志实例并刷新配置")
	allLogMap.Range(func(key string, value *DynamicLogger) bool {
		value.RefreshLogConfig(dataStr)
		return true
	})
}
