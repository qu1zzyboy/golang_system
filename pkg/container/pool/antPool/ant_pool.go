package antPool

import (
	"sync"

	"upbitBnServer/internal/infra/observe/log/staticLog"
	"upbitBnServer/internal/infra/safex"

	"github.com/panjf2000/ants/v2"
)

var (
	cpuPool *ants.Pool
	ioPool  *ants.Pool
	cpuOnce sync.Once
	ioOnce  sync.Once
)

// SubmitToCpuPool 提交任务到计算密集型线程池
func SubmitToCpuPool(protectId string, task func()) error {
	return cpuPool.Submit(safex.SafeGoWrap(protectId, task))
}

func SubmitToIoPool(protectId string, task func()) error {
	return ioPool.Submit(safex.SafeGoWrap(protectId, task))
}

// initCpuPool 初始化全局计算密集型线程池
func initCpuPool(size int) error {
	var initErr error
	cpuOnce.Do(func() {
		pool, err := ants.NewPool(size, ants.WithPreAlloc(true))
		if err != nil {
			initErr = err
			return
		}
		cpuPool = pool
		staticLog.Log.Infof("Ant CPU Pool: %d", size)
	})
	return initErr
}

func initIoPool(size int) error {
	var initErr error
	ioOnce.Do(func() {
		pool, err := ants.NewPool(size, ants.WithPreAlloc(true)) //阻塞提交模式(默认)
		if err != nil {
			initErr = err
			return
		}
		ioPool = pool
		staticLog.Log.Infof("Ant IO Pool: %d", size)
	})
	return initErr
}
