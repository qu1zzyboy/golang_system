package metricIndex

import (
	"context"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/metric"
)

//采样逻辑:
//只有当 Prometheus 访问 /metrics 时才执行,比如每 15 秒一次

var (
	goRoutines     metric.Int64ObservableGauge
	heapAlloc      metric.Int64ObservableGauge
	heapSys        metric.Int64ObservableGauge
	gCCount        metric.Int64ObservableCounter
	gCPauseTotalMs metric.Float64ObservableCounter
)

func RegisterRuntimeMetrics() error {
	var err error
	goRoutines, err = meter.Int64ObservableGauge("runtime.goroutines", metric.WithDescription("当前 goroutine 数"))
	if err != nil {
		return err
	}
	heapAlloc, err = meter.Int64ObservableGauge("runtime.mem.heap_alloc.bytes", metric.WithDescription("堆内存分配(bytes)"))
	if err != nil {
		return err
	}
	heapSys, err = meter.Int64ObservableGauge("runtime.mem.sys.bytes", metric.WithDescription("Go运行时申请的系统内存总量"))
	if err != nil {
		return err
	}
	gCCount, err = meter.Int64ObservableCounter("runtime.gc.count", metric.WithDescription("GC 次数"))
	if err != nil {
		return err
	}
	gCPauseTotalMs, err = meter.Float64ObservableCounter("runtime.gc.pause_total_ms", metric.WithDescription("GC 总暂停时长 (ms)"))
	if err != nil {
		return err
	}
	_, err = meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		o.ObserveInt64(goRoutines, int64(runtime.NumGoroutine()))
		o.ObserveInt64(heapAlloc, int64(m.HeapAlloc))
		o.ObserveInt64(heapSys, int64(m.HeapSys))
		o.ObserveInt64(gCCount, int64(m.NumGC))
		o.ObserveFloat64(gCPauseTotalMs, float64(m.PauseTotalNs)/1e6)
		return nil
	},
		goRoutines, heapAlloc, heapSys, gCCount, gCPauseTotalMs,
	)
	return err
}

// 自定义 Collector：按 scrape 时机读取 runtime 指标
type runtimeCollector struct {
	goroutinesDesc     *prometheus.Desc
	heapAllocDesc      *prometheus.Desc
	heapSysDesc        *prometheus.Desc
	gcCountDesc        *prometheus.Desc
	gcPauseTotalMsDesc *prometheus.Desc
}

func NewRuntimeCollector() *runtimeCollector {
	// 无标签版本；如需按进程/实例打固定标签，可填到 constLabels（第4个参数）
	return &runtimeCollector{
		goroutinesDesc:     prometheus.NewDesc("runtime_goroutines", "当前 goroutine 数", nil, nil),
		heapAllocDesc:      prometheus.NewDesc("runtime_mem_heap_alloc_bytes", "堆内存分配 (bytes)", nil, nil),
		heapSysDesc:        prometheus.NewDesc("runtime_mem_sys_bytes", "Go 运行时向系统申请的内存总量 (bytes)", nil, nil),
		gcCountDesc:        prometheus.NewDesc("runtime_gc_count_total", "GC 次数（累计）", nil, nil),
		gcPauseTotalMsDesc: prometheus.NewDesc("runtime_gc_pause_total_ms", "GC 总暂停时长（毫秒，累计）", nil, nil),
	}
}

func (c *runtimeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.goroutinesDesc
	ch <- c.heapAllocDesc
	ch <- c.heapSysDesc
	ch <- c.gcCountDesc
	ch <- c.gcPauseTotalMsDesc
}

func (c *runtimeCollector) Collect(ch chan<- prometheus.Metric) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	ch <- prometheus.MustNewConstMetric(
		c.goroutinesDesc, prometheus.GaugeValue, float64(runtime.NumGoroutine()),
	)
	ch <- prometheus.MustNewConstMetric(
		c.heapAllocDesc, prometheus.GaugeValue, float64(m.HeapAlloc),
	)
	ch <- prometheus.MustNewConstMetric(
		c.heapSysDesc, prometheus.GaugeValue, float64(m.HeapSys),
	)
	ch <- prometheus.MustNewConstMetric(
		c.gcCountDesc, prometheus.CounterValue, float64(m.NumGC),
	)
	ch <- prometheus.MustNewConstMetric(
		c.gcPauseTotalMsDesc, prometheus.CounterValue, float64(m.PauseTotalNs)/1e6,
	)
}

// 使用自建 Registry,可用这个更灵活的版本
func RegisterRuntimeMetricsWith(reg prometheus.Registerer) error {
	return reg.Register(NewRuntimeCollector())
}
