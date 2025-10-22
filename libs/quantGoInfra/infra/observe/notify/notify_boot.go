package notify

import (
	"context"
)

const MODULE_ID = "notify"

type Boot struct {
	impl Notify
}

func NewBoot(impl_ Notify) *Boot {
	return &Boot{impl: impl_}
}

func (s *Boot) ModuleId() string {
	return MODULE_ID
}

func (s *Boot) DependsOn() []string {
	return []string{}
}

func (s *Boot) Start(ctx context.Context) error {
	setNotify(s.impl)
	return nil
}

func (s *Boot) Stop(ctx context.Context) error {
	return nil
}
