package bnPayloadManagerTest

import (
	"context"

	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/internal/quant/account/accountConfig"
	"upbitBnServer/pkg/singleton"
)

var serviceSingleton = singleton.NewSingleton(func() *Manager {
	return &Manager{}
})

func GetManager() *Manager {
	return serviceSingleton.Get()
}

type Manager struct {
	payload []*Payload // payload处理器
}

func (m *Manager) init(ctx context.Context, read wsDefine.ReadPrivateHandler) error {
	m.payload = make([]*Payload, len(accountConfig.Trades))
	for k, v := range accountConfig.Trades {
		payLoad := newPayload()
		if err := payLoad.init(ctx, v, read); err != nil {
			return err
		}
		m.payload[k] = payLoad
	}
	return nil
}
