package metricIndex

import (
	"github.com/prometheus/client_golang/prometheus"
)

// WebSocket 延迟直方图：标签 stream ∈ {AGG_TRADE, BOOK_TICKER, MARK_PRICE}, phase ∈ {receive, process}
var (
	ReceiveLatencyUS = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "receive_latency_us",
			Help:    "receive WebSocket latency in microseconds",
			Buckets: []float64{500, 1000, 1500, 2000, 2500, 3000, 3500, 4000, 4500, 5000, 10000, 15000, 20000, 200000, 2000000},
		},
		//  stream:BOOK_TICKER
		//  symbol:btcusdt
		//  source:go.binance.c6in.tk
		[]string{"stream", "symbol", "source"},
	)
	ProcessLatencyUS = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "process_latency_us",
			Help:    "process WebSocket latency in microseconds",
			Buckets: []float64{5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 100, 150, 200, 2000, 20000},
		},
		[]string{"stream", "symbol", "source"},
	)
)

// 注册给指定 Registry
func RegisterWSMetrics(reg prometheus.Registerer) {
	reg.MustRegister(ReceiveLatencyUS, ProcessLatencyUS)
}

// ---------------- 观测 API(业务里调用) ----------------

func ObserveReceiveLatencyUS(stream, symbol, source string, us float64) {
	ReceiveLatencyUS.WithLabelValues(stream, symbol, source).Observe(us)
}

func ObserveProcessLatencyUS(stream, symbol, source string, us float64) {
	ProcessLatencyUS.WithLabelValues(stream, symbol, source).Observe(us)
}

// func main() {
// 	reg := prometheus.NewRegistry()
// 	reg.MustRegister(ReceiveLatencyUS)

// 	go func() {
// 		mux := http.NewServeMux()
// 		mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
// 		log.Fatal(http.ListenAndServe(":2112", mux))
// 	}()

// 	// 模拟上报
// 	for {
// 		us := int64(time.Now().UnixNano() % 10000) // 假数据
// 		ReceiveLatencyUS.WithLabelValues("BOOK_TICKER").Observe(float64(us))
// 		time.Sleep(time.Second)
// 	}
// }
