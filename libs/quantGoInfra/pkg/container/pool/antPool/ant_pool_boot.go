package antPool

import (
	"context"

	"github.com/hhh500/quantGoInfra/infra/safex"
)

const (
	MODULE_ID = "antPool"
)

type Boot struct {
	cpuPoolSize, ioPoolSize int
}

func NewBoot(cpuPoolSize, ioPoolSize int) *Boot {
	return &Boot{
		cpuPoolSize: cpuPoolSize,
		ioPoolSize:  ioPoolSize,
	}
}

func (s *Boot) ModuleId() string {
	return MODULE_ID
}

func (s *Boot) DependsOn() []string {
	return []string{
		safex.MODULE_ID,
	}
}

func (s *Boot) Start(ctx context.Context) error {
	if s.cpuPoolSize > 0 {
		if err := initCpuPool(s.cpuPoolSize); err != nil {
			return err
		}
	}
	if s.ioPoolSize > 0 {
		if err := initIoPool(s.ioPoolSize); err != nil {
			return err
		}
	}
	return nil
}

func (s *Boot) Stop(ctx context.Context) error {
	return nil
}
