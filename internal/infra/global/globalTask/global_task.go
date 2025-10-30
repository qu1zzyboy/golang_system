package globalTask

import (
	"sync/atomic"

	"upbitBnServer/internal/infra/observe/log/staticLog"
)

type Service struct {
	name      string
	sourceLen atomic.Int32 //来源计数
	taskArray []func()     //多个来源的实现
}

func (m *Service) RegisterTask(index uint8, task func()) {
	if task == nil {
		return
	}

	if m.taskArray[index] == nil {
		m.sourceLen.Add(1)
		m.taskArray[index] = task
	}
	staticLog.Log.Infof("RegisterTask[%s]: index=%d, task=%p, len=%d", m.name, index, task, m.sourceLen.Load())
}

func (m *Service) RunAll() {
	if m.sourceLen.Load() == 0 {
		return
	}
	for _, task := range m.taskArray {
		if task != nil {
			task()
		}
	}
}
