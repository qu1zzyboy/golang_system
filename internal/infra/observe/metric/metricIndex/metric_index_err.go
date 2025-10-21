package metricIndex

import (
	"go.opentelemetry.io/otel/metric"
)

//采样逻辑
//手动在代码中调用 .Add(ctx, value)

var (
	ErrorTotal metric.Int64Counter
	PanicTotal metric.Int64Counter
)

func RegisterErrorMetrics() error {
	var err error
	ErrorTotal, err = meter.Int64Counter("error_total", metric.WithDescription("全局错误汇总"))
	if err != nil {
		return err
	}
	PanicTotal, err = meter.Int64Counter("panic_total", metric.WithDescription("全局panic汇总"))
	return err
}
