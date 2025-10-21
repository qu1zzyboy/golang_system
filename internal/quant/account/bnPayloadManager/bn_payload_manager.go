package bnPayloadManager

import (
	"context"

	"github.com/hhh500/quantGoInfra/pkg/singleton"
	"github.com/hhh500/upbitBnServer/internal/quant/account/accountConfig"
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

func (m *Manager) init(ctx context.Context) error {
	m.payload = make([]*Payload, len(accountConfig.Trades))
	for k, v := range accountConfig.Trades {
		payLoad := newPayload()
		if err := payLoad.init(ctx, v); err != nil {
			return err
		}
		m.payload[k] = payLoad
	}
	return nil
}
