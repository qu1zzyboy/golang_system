package bnOrderAppManager

import (
	"context"

	"upbitBnServer/internal/quant/account/accountConfig"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/pkg/singleton"
)

var monitorSingleton = singleton.NewSingleton(func() *MonitorManager {
	return &MonitorManager{}
})

func GetMonitorManager() *MonitorManager {
	return monitorSingleton.Get()
}

type MonitorManager struct {
	appArray []*OrderApp // payload处理器
}

func (m *MonitorManager) init(ctx context.Context) error {
	m.appArray = make([]*OrderApp, len(accountConfig.Monitors))
	for k, v := range accountConfig.Monitors {
		app := newOrderApp()
		if err := app.Init(ctx, v); err != nil {
			return err
		}
		app.isMonitor = true
		m.appArray[k] = app
	}
	return nil
}

func (m *MonitorManager) SendMonitorOrder(index uint8, req orderModel.MyPlaceOrderReq) error {
	return m.appArray[index].wsOrder.CreateOrder(req)
}
