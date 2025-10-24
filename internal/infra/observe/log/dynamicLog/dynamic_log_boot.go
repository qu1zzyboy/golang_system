package dynamicLog

import (
	"context"

	"upbitBnServer/internal/infra/global/globalCron"
	"upbitBnServer/internal/infra/global/globalTask"
	"upbitBnServer/pkg/utils/timeUtils"
)

const (
	MODULE_ID = "dynamic_log" //全局定时任务
)

type Boot struct {
}

func NewBoot() *Boot {
	return &Boot{}
}

func (s *Boot) ModuleId() string {
	return MODULE_ID
}

func (s *Boot) DependsOn() []string {
	return []string{globalCron.MODULE_ID}
}

func (s *Boot) Start(ctx context.Context) error {
	globalTask.GetDayBeginService().RegisterTask(uint8(globalTask.RefreshLogFile), func() { refreshAllLog(timeUtils.GetNowDateStr()) })
	return nil
}

func (s *Boot) Stop(ctx context.Context) error {
	return nil
}
