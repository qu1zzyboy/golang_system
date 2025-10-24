package bnOrderAppManager

import (
	"context"

	"upbitBnServer/internal/quant/execute/order/orderSdk/bn/orderSdkBnModel"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
)

func (s *TradeManager) SetBnLeverage(leverage uint8, symbolName string) error {
	ctx := context.Background()
	for k, app := range s.appArray {
		res, err := app.rest.DoLeverage(ctx, &orderSdkBnModel.FutureLeverageSdk{
			Symbol:   symbolName,
			Leverage: leverage,
		})
		if err != nil {
			toUpBitListDataStatic.DyLog.GetLog().Errorf("[%d]设置币安[%s]杠杆失败,leverage:%d,err:%v", k, symbolName, leverage, err)
			return err
		}
		toUpBitListDataStatic.DyLog.GetLog().Infof("[%d]设置币安[%s]杠杆成功,leverage:%d,res:%v", k, symbolName, leverage, res)
	}
	return nil
}
