package bnOrderAppManager

import (
	"context"
	"upbitBnServer/internal/strategy/newsDrive/bn/bnDriverWsParse"
	"upbitBnServer/server/instanceEnum"

	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/quant/account/accountConfig"
	"upbitBnServer/internal/quant/execute/order/orderSdk/bn/orderSdkBnRest"
	"upbitBnServer/internal/quant/execute/order/orderSdk/bn/orderSdkBnWsSign"
	"upbitBnServer/internal/quant/execute/order/wsRequestCache"

	"github.com/tidwall/gjson"
)

type OrderApp struct {
	rest         *orderSdkBnRest.FutureRest     // REST API 客户端
	wsOrder      *orderSdkBnWsSign.FutureClient // WS API 客户端
	accountKeyId uint8                          // 账户序号
	isMonitor    bool                           // 是否是监控账户
}

func newOrderApp() *OrderApp {
	return &OrderApp{}
}

func (s *OrderApp) init(ctx context.Context, v accountConfig.Config) error {
	s.accountKeyId = v.AccountId
	s.rest = orderSdkBnRest.NewFutureRest(v.ApiKeyHmac, v.SecretHmac)
	s.wsOrder = orderSdkBnWsSign.NewFutureClient(v.ApiKeyHmac, v.SecretHmac)
	if err := s.wsOrder.RegisterReadHandler(ctx, v.AccountId, s.OnWsOrder); err != nil {
		return err
	}
	return nil
}

func (s *OrderApp) OnWsOrder(data []byte) {
	reqId := systemx.WsId16B(gjson.GetBytes(data, "id").String())
	reqOk := gjson.GetBytes(data, "status").Int() == 200

	wsMeta, ok := wsRequestCache.GetCache().GetMeta(reqId)
	if !ok {
		// 帐号认证的json
		dynamicLog.Error.GetLog().Errorf("[%d]WS_REQUEST: [%s]  not found %s", s.accountKeyId, string(reqId[:]), string(data))
		return
	}

	// 接受到返回就删除 wsRequestCache
	defer wsRequestCache.GetCache().DelMeta(reqId)

	switch wsMeta.ReqFrom {
	case instanceEnum.DRIVER_LIST_BN:
		bnDriverWsParse.Get().Parse(data, wsMeta, reqId, s.accountKeyId, reqOk)
	case instanceEnum.TEST:
	default:
		dynamicLog.Error.GetLog().Errorf("WS_REQUEST: unknown ReqFrom %s", instanceEnum.String(wsMeta.ReqFrom))
	}
}
