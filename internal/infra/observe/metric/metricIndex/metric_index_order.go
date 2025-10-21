package metricIndex

import (
	"go.opentelemetry.io/otel/metric"
)

var (
	OrdersTotal  metric.Int64Counter
	OrderLatency metric.Float64Histogram
	OrdersFailed metric.Int64Counter
)

func RegisterOrderMetrics() error {
	var err error
	OrdersTotal, err = meter.Int64Counter("orders_total", metric.WithDescription("总订单数"))
	if err != nil {
		return err
	}
	OrderLatency, err = meter.Float64Histogram("order_latency_ms", metric.WithDescription("下单延迟"))
	if err != nil {
		return err
	}
	OrdersFailed, err = meter.Int64Counter("orders_failed_total", metric.WithDescription("下单失败统计"))
	if err != nil {
		return err
	}
	return nil
}
