package toUpBitListBnAccount

import (
	"context"
	"errors"

	"upbitBnServer/internal/quant/account/accountConfig"
	"upbitBnServer/internal/quant/account/universalTransfer"
	"upbitBnServer/internal/quant/execute/order/orderSdk/bn/orderSdkBnModel"
	"upbitBnServer/internal/quant/execute/order/orderSdk/bn/orderSdkBnRest"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitBnMode"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/pkg/singleton"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

const (
	email = "chunpingzhan888@gmail.com"
)

var (
	bnSingleton = singleton.NewSingleton(func() *BnAccountManager {
		return &BnAccountManager{
			sp: orderSdkBnRest.NewSpotRest("j0G54QHYH68pETkd0K9keDW2U02woj9w5mJ8EoFrv8tlRi0fCzp8XjwTXqFVNtho", "F7HmmiZnDUpgVgOW0CAFrVn7FcqFUN8t9UbYxQRKqZlIURE9sZiy7hqhXbkMIOC2"),
			fu: orderSdkBnRest.NewFutureRest("j0G54QHYH68pETkd0K9keDW2U02woj9w5mJ8EoFrv8tlRi0fCzp8XjwTXqFVNtho", "F7HmmiZnDUpgVgOW0CAFrVn7FcqFUN8t9UbYxQRKqZlIURE9sZiy7hqhXbkMIOC2"),
		}
	})
)

func GetBnAccountManager() *BnAccountManager {
	return bnSingleton.Get()
}

type BnAccountManager struct {
	sp *orderSdkBnRest.SpotRest   // 不参与交易,只用来查询资金
	fu *orderSdkBnRest.FutureRest //
}

func (s *BnAccountManager) TransferIn(from int32, amount decimal.Decimal) error {
	out := accountConfig.Trades[from]
	usdtAmount := toUpbitBnMode.Mode.GetTransferAmount(amount)
	if usdtAmount.LessThan(decimal.Zero) {
		toUpBitDataStatic.DyLog.GetLog().Infof("账户[%d]余额不足200u,不划转", out.AccountId)
		return nil
	}
	var reqIn universalTransfer.UniversalTransferReq
	reqIn.To = email
	reqIn.From = out.Email
	reqIn.FromAcType = string(universalTransfer.USDT_FUTURE)
	reqIn.ToAcType = string(universalTransfer.USDT_FUTURE)
	reqIn.Asset = "USDT"
	reqIn.Amount = usdtAmount.Truncate(0)
	resIn, reqParamIn, err := s.sp.DoTransfer(context.Background(), orderSdkBnModel.GetSpotTransferSdk(&reqIn))
	if err != nil {
		toUpBitDataStatic.DyLog.GetLog().Errorf("[%d]划转到母账户[%s]失败,err:%s,请求参数:%s", out.AccountId, reqIn.ToAcType, err.Error(), reqParamIn)
		return err
	}
	toUpBitDataStatic.DyLog.GetLog().Infof("账户[%d]划转[%s,%s]usdt到母账户[%s]成功:%s", out.AccountId, amount, reqIn.Amount, reqIn.ToAcType, resIn)
	return nil
}

func (s *BnAccountManager) TransferOut(to int32, amount decimal.Decimal) error {
	in := accountConfig.Trades[to]
	var reqOut universalTransfer.UniversalTransferReq
	reqOut.From = email
	reqOut.To = in.Email
	reqOut.FromAcType = string(universalTransfer.USDT_FUTURE)
	reqOut.ToAcType = string(universalTransfer.USDT_FUTURE)
	reqOut.Asset = "USDT"
	reqOut.Amount = amount.Truncate(0)
	resOut, reqParamOut, err := s.sp.DoTransfer(context.Background(), orderSdkBnModel.GetSpotTransferSdk(&reqOut))
	if err != nil {
		toUpBitDataStatic.DyLog.GetLog().Errorf("母账户划转到[%d]失败,err:%s,请求参数:%s", in.AccountId, err.Error(), reqParamOut)
		return err
	}
	if gjson.Get(resOut, "tranId").Exists() {
		toUpBitDataStatic.DyLog.GetLog().Infof("母账户划转[%s]usdt到[%d][%s]成功:%s", reqOut.Amount.String(), in.AccountId, reqOut.ToAcType, resOut)
		return nil
	}
	toUpBitDataStatic.DyLog.GetLog().Infof("母账户划转[%s]usdt到[%d][%s]失败:%s", reqOut.Amount.String(), in.AccountId, reqOut.ToAcType, resOut)
	return errors.New(resOut)
}

func (s *BnAccountManager) RefreshSymbolConfig() error {
	data, err := s.fu.DoSymbolConfig(context.Background(), orderSdkBnModel.NewFutureSymbolConfigSdk())
	if err != nil {
		return err
	}
	// 遍历整个数组
	gjson.ParseBytes(data).ForEach(func(key, value gjson.Result) bool {
		symbolIndex, ok := toUpBitDataStatic.SymbolIndex.Load(value.Get("symbol").String())
		if !ok {
			return true
		}
		toUpBitDataStatic.SymbolMaxNotional.Store(symbolIndex, decimal.RequireFromString(value.Get("maxNotionalValue").String()))
		return true
	})
	toUpBitDataStatic.DyLog.GetLog().Infof("共刷新%d个交易对开仓上限信息", toUpBitDataStatic.SymbolMaxNotional.Length())
	return nil
}

// {"code":-9000,"msg":"user have no available amount"}
// {"tranId":313310550555}
