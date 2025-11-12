package bnOrderAppManager

import (
	"context"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/instanceEnum"
	"upbitBnServer/internal/quant/execute/order/orderSdk/bn/orderSdkBnWsNoSign"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitBnWsParse"

	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/quant/account/accountConfig"
	"upbitBnServer/internal/quant/execute/order/orderSdk/bn/orderSdkBnRest"
	"upbitBnServer/internal/quant/execute/order/wsRequestCache"
)

type OrderApp struct {
	rest         *orderSdkBnRest.FutureRest       // REST API 客户端
	wsOrder      *orderSdkBnWsNoSign.FutureClient // WS API 客户端
	accountKeyId uint8                            // 账户序号
	isMonitor    bool                             // 是否是监控账户
}

func newOrderApp() *OrderApp {
	return &OrderApp{}
}

func (s *OrderApp) Init(ctx context.Context, v accountConfig.Config) error {
	s.accountKeyId = v.AccountId
	s.rest = orderSdkBnRest.NewFutureRest(v.ApiKeyHmac, v.SecretHmac)
	if v.ApiKeyEd25519 != "" {
		s.wsOrder = orderSdkBnWsNoSign.NewFutureClient(v.ApiKeyEd25519, v.SecretEd25519)
		if err := s.wsOrder.RegisterReadHandler(ctx, v.AccountId, s.OnWsResp); err != nil {
			return err
		}
	}
	return nil
}

func (s *OrderApp) OnWsResp(data []byte) {
	var reqId systemx.WsId16B
	copy(reqId[:], data[7:23])

	wsMeta, ok := wsRequestCache.GetCache().GetMeta(reqId)
	if !ok {
		// 帐号认证的json
		if data[34] == '2' && data[35] == '0' && data[36] == '0' {
			return
		}
		dynamicLog.Error.GetLog().Errorf("[%d]WS_REQUEST: [%s]  not found %s", s.accountKeyId, string(reqId[:]), string(data))
		return
	}
	// 最终删除ws_meta
	defer wsRequestCache.GetCache().DelMeta(reqId)

	switch wsMeta.ReqFrom {
	case instanceEnum.TO_UPBIT_LIST_BN:
		toUpbitBnWsParse.Get().Parse(data, wsMeta, reqId, s.accountKeyId)
	case instanceEnum.TEST:
	default:
		dynamicLog.Error.GetLog().Errorf("WS_REQUEST: unknown ReqFrom %s", wsMeta.ReqFrom.String())
	}
}
