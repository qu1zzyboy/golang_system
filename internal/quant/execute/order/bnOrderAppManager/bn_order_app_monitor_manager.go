package bnOrderAppManager

import (
	"context"

	"github.com/hhh500/quantGoInfra/pkg/singleton"
	"github.com/hhh500/upbitBnServer/internal/quant/account/accountConfig"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderModel"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderSdk/bn/orderSdkBnModel"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderStatic"
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
	err := m.appArray[index].wsOrderSign.CreateOrder(reqFrom, orderSdkBnModel.GetFuturePlaceLimitSdk(req))
	if err == nil {
		orderStatic.GetService().SaveOrderMeta(req.ClientOrderId, orderStatic.StaticMeta{
			SymbolIndex: symbolIndex,
			OrderMode:   req.OrderMode,
			OrderFrom:   reqFrom,
		})
	}
	return err
}
