package globalTask

import (
	"sync/atomic"

	"github.com/hhh500/quantGoInfra/pkg/singleton"
)

type Min10EndRun uint8

const (
	SaveWsData Min10EndRun = iota
	TOTAL_Min10_COUNT
)

var (
	min10Singleton = singleton.NewSingleton(func() *Service {
		return &Service{
			name:      "min10_end_run",
			sourceLen: atomic.Int32{},
			taskArray: make([]func(), TOTAL_Min10_COUNT),
		}
	})
)

func GetMin10EndService() *Service {
	return min10Singleton.Get()
}
