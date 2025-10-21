package metricx

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

type Config struct {
	Enabled     bool   // 是否启用 metrics
	ServiceName string // 用于 Prometheus 标签中的服务标识
	Port        string // 用于启动 metrics HTTP 接口,例如 ":9090"
}

func Init(cfg Config) error {
	// 创建 Prometheus exporter
	exp, err := prometheus.New()
	if err != nil {
		return err
	}
	// 创建 Meter Provider
	provider := metric.NewMeterProvider(
		metric.WithReader(exp),
		metric.WithResource(resource.NewSchemaless(
			semconv.ServiceName(cfg.ServiceName),
		)),
	)

	// 设置全局 Provider
	otel.SetMeterProvider(provider)

	go func() {
		// 启动 HTTP 服务,暴露 /metrics 路由
		http.Handle("/metrics", promhttp.Handler())
		log.Println("开始监听 :2112/metrics")
		if err := http.ListenAndServe(":2112", nil); err != nil {
			log.Fatalf("HTTP 服务启动失败: %v", err)
		}
	}()
	return nil
}

// ----------------------------
// ✅ Prometheus exporter 初始化
// ✅ 全局 MeterProvider 注入
// ✅ 支持 /metrics HTTP 暴露
// ✅ 配合 trace/log 构成 observability 三大件
