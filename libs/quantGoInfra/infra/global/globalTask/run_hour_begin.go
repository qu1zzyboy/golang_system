package globalTask

import (
	"sync/atomic"

	"github.com/hhh500/quantGoInfra/pkg/singleton"
)

type HourBeginRun uint8

const (
	RefreshConfigFile HourBeginRun = iota // RefreshConfigFile 每小时刷新配置文件
	RefreshSymbolInfo                     // RefreshSymbolInof 每小时刷新交易对信息
	TOTAL_HOUR_COUNT                      // TOTAL_COUNT 总数
)

var hourSingleton = singleton.NewSingleton(func() *Service {
	return &Service{
		name:      "hour_begin_run",
		sourceLen: atomic.Int32{},
		taskArray: make([]func(), TOTAL_HOUR_COUNT),
	}
})

func GetHourBeginService() *Service {
	return hourSingleton.Get()
}
