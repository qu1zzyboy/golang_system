package orderSdkBnWsSign

import (
	"github.com/hhh500/quantGoInfra/pkg/container/pool/byteBufPool"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderSdk/bn/orderSdkBnModel"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/wsRequestCache"
)

func (s *FutureClient) CancelOrder(reqFrom orderBelongEnum.Type, api *orderSdkBnModel.FutureQuerySdk) error {
	rawData, err := api.ParseWsReqFast("C", s.apiKey, "order.cancel", s.secretByte)
	defer byteBufPool.ReleaseBuffer(rawData)
	if rawData == nil || err != nil {
		return err
	}
	if err = s.conn.WriteAsync(*rawData); err != nil {
		return err
	}
	wsRequestCache.GetCache().StoreMeta("C"+api.ClientOrderId, &wsRequestCache.WsRequestMeta{
		Json:    string(*rawData),
		ReqType: wsRequestCache.CANCEL_ORDER,
		ReqFrom: reqFrom,
	})
	return err
}
