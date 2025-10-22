package globalTask

import (
	"sync/atomic"

	"github.com/hhh500/quantGoInfra/pkg/singleton"
)

type DayBeginRun uint8

const (
	RefreshLogFile DayBeginRun = iota // RefreshLogFile 每天刷新日志文件
	TOTAL_DAY_COUNT
)

var (
	daySingleton = singleton.NewSingleton(func() *Service {
		return &Service{
			name:      "day_begin_run",
			sourceLen: atomic.Int32{},
			taskArray: make([]func(), TOTAL_DAY_COUNT),
		}
	})
)

func GetDayBeginService() *Service {
	return daySingleton.Get()
}
