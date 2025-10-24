package registerHandler

import (
	"context"

	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/infra/errorx/errDefine"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/pkg/container/map/myMap"
	"upbitBnServer/pkg/utils/convertx"
	"upbitBnServer/server/serverInstanceEnum"
)

/***
这是一个通用的注册、回调组件
***/

const (
	REGISTER    = "registerHandler"
	REGISTER_OR = "registerOrReplace"
	UNREGISTER  = "unregisterHandler"
)

type Registry[T any] struct {
	handlers myMap.MySyncMap[serverInstanceEnum.Type, T]
}

// NewRegistry 构造函数,创建一个带日志和监控的注册中心
func NewRegistry[T any]() *Registry[T] {
	return &Registry[T]{handlers: myMap.NewMySyncMap[serverInstanceEnum.Type, T]()}
}

// Register 注册策略 handler,不允许重复注册
func (r *Registry[T]) Register(ctx context.Context, instanceId serverInstanceEnum.Type, fields map[string]string, handler T) error {
	if _, ok := r.handlers.Load(instanceId); ok {
		return errDefine.InstanceExists.WithMetadata(fields)
	}
	r.handlers.Store(instanceId, handler)
	r.reportSuccess(ctx, "策略注册", fields)
	return nil
}

// RegisterOrReplace 注册策略,如果已存在则覆盖
func (r *Registry[T]) RegisterOrReplace(ctx context.Context, instanceId serverInstanceEnum.Type, fields map[string]string, handler T) error {
	r.handlers.Store(instanceId, handler)
	r.reportSuccess(ctx, "策略注册或更新", fields)
	return nil
}

// Unregister 删除已注册策略
func (r *Registry[T]) Unregister(ctx context.Context, instanceId serverInstanceEnum.Type, fields map[string]string) error {
	r.handlers.Delete(instanceId)
	r.reportSuccess(ctx, "策略注销", fields)
	return nil
}

// Get 获取策略 handler(只读)
func (r *Registry[T]) Get(instanceId serverInstanceEnum.Type) (T, bool) {
	return r.handlers.Load(instanceId)
}

// Range 遍历所有策略
func (r *Registry[T]) Range(f func(serverInstanceEnum.Type, T) bool) {
	r.handlers.Range(func(key serverInstanceEnum.Type, value T) bool {
		return f(key, value)
	})
}

func (r *Registry[T]) Count() int {
	return r.handlers.Length()
}

// 内部上报逻辑
func (r *Registry[T]) reportSuccess(ctx context.Context, msg string, fields map[string]string) {
	fields[defineJson.RefCount] = convertx.ToString(r.Count())
	dynamicLog.Log.GetLog().Info(msg, " ", fields)
}
