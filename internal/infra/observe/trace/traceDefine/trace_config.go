package traceDefine

type ExporterType string

const (
	ExporterOtlpGrpc ExporterType = "otlp_grpc"
	ExporterOtlpHttp ExporterType = "otlp_http"
)

type Config struct {
	Enabled      bool         // 是否启用 trace
	SampleRate   float64      //采样比例(0.0~1.0),控制性能与粒度
	ServiceName  string       // 用于 trace 资源标识
	ExporterType ExporterType //输出方式(比如到 Jaeger、OTLP、终端打印)
	EndPoint     string       // 导出目标地址,如 "localhost:4317"、Jaeger HTTP 服务
}
