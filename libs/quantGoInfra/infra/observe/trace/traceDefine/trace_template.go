package traceDefine

import (
	"context"

	"github.com/hhh500/quantGoInfra/define/defineJson"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer = otel.Tracer(defineJson.QuantSystem) //一个打点工具：用来创建和管理 span
)

func GetTracer() trace.Tracer {
	return tracer
}

// TraceSpanDef 表示一个可复用的 Trace Span 模板定义
// 场景:大量重复的 trace定义
type TraceSpanDef struct {
	Name   string
	Fields []attribute.KeyValue
}

// Start 从这个模板生成一个新的 span(带有名称和属性)
func (t TraceSpanDef) Start(ctx context.Context) (context.Context, trace.Span) {
	ctx, span := tracer.Start(ctx, t.Name)
	span.SetAttributes(t.Fields...)
	return ctx, span
}
