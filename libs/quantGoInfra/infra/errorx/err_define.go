package errorx

import (
	"errors"
	"fmt"
	"maps"

	"github.com/hhh500/quantGoInfra/define/defineJson"
	"github.com/hhh500/quantGoInfra/infra/debugx"
	"github.com/hhh500/quantGoInfra/infra/errorx/errCode"
	"github.com/hhh500/quantGoInfra/pkg/utils/convertx"
	"go.opentelemetry.io/otel/attribute"
)

// Error 是通用错误结构,支持链式错误与附加信息
type Error struct {
	Code     uint16            // 错误码唯一标识
	Message  string            // 用户可读信息
	Metadata map[string]string // 附加元信息
	cause    error             // 底层错误链
}

func (e *Error) Error() string {
	return fmt.Sprintf("error: code = %d,  message = %s, metadata = %v, cause = %v", e.Code, e.Message, e.Metadata, e.cause)
}

func (e *Error) Unwrap() error {
	return e.cause
}

func (e *Error) Is(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == e.Code
	}
	return false
}

func (e *Error) WithCause(cause error) *Error {
	err := clone(e)
	err.cause = cause
	return err
}

func (e *Error) WithMetadata(meta map[string]string) *Error {
	err := clone(e)
	maps.Copy(err.Metadata, meta)
	// 只有当 Caller 不存在时,才设置
	if _, exists := err.Metadata[defineJson.Caller]; !exists {
		err.Metadata[defineJson.Caller] = debugx.GetCaller(2)
	}
	return err
}

func (e *Error) Attrs() []attribute.KeyValue {
	if e == nil {
		return nil
	}
	attrs := []attribute.KeyValue{
		attribute.Int(defineJson.ErrCode, int(e.Code)),
		attribute.String(defineJson.ErrMsg, e.Message),
	}
	for k, v := range e.Metadata {
		attrs = append(attrs, attribute.String(k, v))
	}
	return attrs
}

func (e *Error) Fields() map[string]string {
	if e == nil {
		return nil
	}
	fields := map[string]string{
		defineJson.ErrCode: convertx.ToString(e.Code),
		defineJson.ErrMsg:  e.Message,
	}
	maps.Copy(fields, e.Metadata)
	return fields
}

func clone(err *Error) *Error {
	if err == nil {
		return nil
	}
	metadata := map[string]string{}
	if err.Metadata != nil {
		metadata = make(map[string]string, len(err.Metadata))
		maps.Copy(metadata, err.Metadata)
	}
	return &Error{
		cause:    err.cause,
		Code:     err.Code,
		Message:  err.Message,
		Metadata: metadata,
	}
}

func New(code uint16, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func Newf(code uint16, format string, a ...any) *Error {
	return New(code, fmt.Sprintf(format, a...))
}

func Code(err error) uint16 {
	if err == nil {
		return errCode.CODE_SUCCESS
	}
	return FromError(err).Code
}

func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return New(errCode.CODE_UN_KNOWN, err.Error())
}

// TryAddMetadata 尝试将 metadata 附加到 error 中,如果 error 是 *Error
func TryAddMetadata(err error, meta map[string]string) error {
	var e *Error
	if errors.As(err, &e) {
		return e.WithMetadata(meta)
	}
	return err
}
