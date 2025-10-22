package tracex

import (
	"context"

	"github.com/hhh500/quantGoInfra/infra/observe/trace/traceDefine"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// WithTrace 当前这一段业务逻辑 / 操作 / 步骤 的命名
func WithTrace(ctx context.Context, isLeaf bool, spanName string, fn func(ctx context.Context) error, attrs ...attribute.KeyValue) error {
	// 1. 创建 span(挂到原 ctx 上)
	ctx, span := traceDefine.GetTracer().Start(ctx, spanName)
	defer span.End()

	// 2. 设置属性
	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}

	// 3. 执行业务逻辑
	err := fn(ctx)

	// 4. 如果是叶子节点,记录错误和状态
	if isLeaf {
		if err != nil {
			span.RecordError(err)                    // 记录错误到span logs
			span.SetStatus(codes.Error, err.Error()) // 记录错误到span tags,会设置error=true,otel.status_code=ERROR,otel.status_description="错误信息"
		}
	}
	return err
}

// AddTraceEvent 向当前 Span 添加结构化事件
func AddTraceEvent(ctx context.Context, spanName string, attrs ...attribute.KeyValue) {
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		span.AddEvent(spanName, trace.WithAttributes(attrs...))
	}
}
