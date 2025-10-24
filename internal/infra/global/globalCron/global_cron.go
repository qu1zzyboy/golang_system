package globalCron

import (
	"github.com/robfig/cron/v3"
)

func initDefault() {
	once.Do(func() {
		c = cron.New(cron.WithSeconds())
		c.Start()
	})
}

// AddFunc 注册任务
func AddFunc(spec string, cmd func()) (cron.EntryID, error) {
	initDefault() // 确保 cron 启动
	return c.AddFunc(spec, cmd)
}

func RemoveFunc(id cron.EntryID) {
	initDefault() // 确保 cron 启动
	c.Remove(id)
}

// Stop 可选:停止调度器
func Stop() {
	if c != nil {
		c.Stop()
	}
}
