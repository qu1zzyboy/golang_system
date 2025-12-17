package bnOrderAppManager

import (
	"context"

	"upbitBnServer/internal/quant/account/accountConfig"
	"upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/quant/execute/order/orderSdk/bn/orderSdkBnModel"
	"upbitBnServer/pkg/singleton"
)

var tradeSingleton = singleton.NewSingleton(func() *TradeManager {
	return &TradeManager{}
})

func GetTradeManager() *TradeManager {
	return tradeSingleton.Get()
}

type TradeManager struct {
	appArray []*OrderApp
}

func (s *TradeManager) init(ctx context.Context) error {
	s.appArray = make([]*OrderApp, len(accountConfig.Trades))
	for k, v := range accountConfig.Trades {
		app := newOrderApp()
		if err := app.init(ctx, v); err != nil {
			return err
		}
		s.appArray[k] = app
	}
	return nil
}

func (s *TradeManager) SendPlaceOrder(index uint8, symbolIndex int, req *orderModel.MyPlaceOrderReq) error {
	return s.appArray[index].wsOrder.CreateOrder(req)
}

func (s *TradeManager) SendCancelOrder(index uint8, req *orderModel.MyQueryOrderReq) error {
	return s.appArray[index].wsOrder.CancelOrder(req)
}

func (s *TradeManager) SendModifyOrder(reqFrom orderBelongEnum.Type, index uint8, req *orderModel.MyModifyOrderReq) error {
	return s.appArray[index].wsOrder.ModifyOrder(reqFrom, orderSdkBnModel.GetFutureModifySdk(req))
}

func (s *TradeManager) SendQueryAccountBalance(reqFrom orderBelongEnum.Type, index uint8) error {
	return s.appArray[index].wsOrder.QueryAccount(reqFrom)
}
