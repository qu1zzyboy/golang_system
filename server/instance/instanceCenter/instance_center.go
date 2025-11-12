package instanceCenter

import (
	"context"

	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/infra/errorx/errDefine"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/systemx/instanceEnum"
	"upbitBnServer/internal/resource/registerHandler"
	"upbitBnServer/pkg/singleton"
	"upbitBnServer/pkg/utils/convertx"
	"upbitBnServer/server/instance"
)

type Manager struct {
	handlers *registerHandler.Registry[instance.Instance]
}

var serviceSingleton = singleton.NewSingleton(func() *Manager {
	return &Manager{
		handlers: registerHandler.NewRegistry[instance.Instance](),
	}
})

func GetManager() *Manager {
	return serviceSingleton.Get()
}

func (s *Manager) Register(ctx context.Context, instanceId instanceEnum.Type, fields map[string]string, handler instance.Instance) error {
	fields[defineJson.From] = "instanceCenter"
	return s.handlers.RegisterOrReplace(ctx, instanceId, fields, handler)
}

func (s *Manager) UnRegister(ctx context.Context, instanceId instanceEnum.Type, fields map[string]string) error {
	return s.handlers.Unregister(ctx, instanceId, fields)
}

func (s *Manager) StopInstance(ctx context.Context, instanceId instanceEnum.Type) error {
	if handler, ok := s.handlers.Get(instanceId); ok {
		return handler.OnStop(ctx)
	}
	return errDefine.InstanceNotExists.WithMetadata(map[string]string{defineJson.InstanceId: convertx.ToString(uint8(instanceId))})
}

func (s *Manager) UpdateInstance(ctx context.Context, instanceId instanceEnum.Type, param instance.InstanceUpdate) error {
	if handler, ok := s.handlers.Get(instanceId); ok {
		return handler.OnUpdate(ctx, param)
	}
	return errDefine.InstanceNotExists.WithMetadata(map[string]string{defineJson.InstanceId: convertx.ToString(uint8(instanceId))})
}

func (s *Manager) IsInstanceExists(instanceId instanceEnum.Type) bool {
	_, ok := s.handlers.Get(instanceId)
	return ok
}

func (s *Manager) PrintAll() {
	dynamicLog.Log.GetLog().Infof("当前所有实例信息:[%d]", s.handlers.Count())
	s.handlers.Range(func(instanceId instanceEnum.Type, handler instance.Instance) bool {
		dynamicLog.Log.GetLog().Infof("Instance ID: %s, Handler Type: %T", instanceId, handler)
		return true
	})
}
