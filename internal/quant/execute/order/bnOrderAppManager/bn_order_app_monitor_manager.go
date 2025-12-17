package bnOrderAppManager

import (
	"context"

	"upbitBnServer/internal/quant/account/accountConfig"
	"upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/quant/execute/order/orderSdk/bn/orderSdkBnModel"
	"upbitBnServer/internal/quant/execute/order/orderStaticMeta"
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
		if err := app.init(ctx, v); err != nil {
			return err
		}
		app.isMonitor = true
		m.appArray[k] = app
	}
	return nil
}

func (m *MonitorManager) SendMonitorOrder(reqFrom orderBelongEnum.Type, index uint8, symbolIndex int, req *orderModel.MyPlaceOrderReq) error {
	err := m.appArray[index].wsOrder.CreateOrder(reqFrom, orderSdkBnModel.GetFuturePlaceLimitSdk(req))
	if err == nil {
		orderStaticMeta.GetService().SaveOrderMeta(req.ClientOrderId, orderStaticMeta.StaticMeta{
			SymbolIndex: symbolIndex,
			OrderMode:   req.OrderMode,
			OrderFrom:   reqFrom,
		})
	}
	return err
}
