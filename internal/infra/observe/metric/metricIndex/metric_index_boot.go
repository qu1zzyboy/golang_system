package metricIndex

import (
	"context"

	"upbitBnServer/internal/define/defineJson"

	"go.opentelemetry.io/otel"
)

const (
	ModuleId = "metricIndex"
)

var (
	meter = otel.Meter(defineJson.QuantSystem)
)

type Boot struct {
}

func NewBoot() *Boot {
	return &Boot{}
}

func (s *Boot) ModuleId() string {
	return ModuleId
}

func (s *Boot) DependsOn() []string {
	return []string{}
}

func (s *Boot) Start(ctx context.Context) error {
	if err := RegisterErrorMetrics(); err != nil {
		return err
	}
	if err := RegisterOrderMetrics(); err != nil {
		return err
	}
	if err := RegisterRuntimeMetrics(); err != nil {
		return err
	}
	return nil
}

func (s *Boot) Stop(ctx context.Context) error {
	return nil
}
