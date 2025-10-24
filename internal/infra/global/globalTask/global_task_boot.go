package globalTask

import (
	"context"

	"upbitBnServer/internal/define/defineTime"
	"upbitBnServer/internal/infra/global/globalCron"
)

const (
	MODULE_ID = "global_task" //全局定时任务
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
	var err error
	_, err = globalCron.AddFunc(defineTime.Min10EndStr, func() {
		GetMin10EndService().RunAll()
	})
	if err != nil {
		return err
	}
	_, err = globalCron.AddFunc(defineTime.HourBegin6MStr, func() {
		GetHourBeginService().RunAll()
	})
	if err != nil {
		return err
	}
	_, err = globalCron.AddFunc(defineTime.DayBeginStr, func() {
		GetDayBeginService().RunAll()
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Boot) Stop(ctx context.Context) error {
	return nil
}
