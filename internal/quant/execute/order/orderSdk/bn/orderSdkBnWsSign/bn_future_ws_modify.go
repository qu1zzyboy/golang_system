package orderSdkBnWsSign

import (
	"upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"upbitBnServer/internal/quant/execute/order/orderSdk/bn/orderSdkBnModel"
	"upbitBnServer/internal/quant/execute/order/wsRequestCache"
	"upbitBnServer/pkg/container/pool/byteBufPool"
)

func (s *FutureClient) ModifyOrder(reqFrom orderBelongEnum.Type, api *orderSdkBnModel.FutureModifySdk) error {
	rawData, err := api.ParseWsReqFast(s.apiKey, s.secretByte)
	defer byteBufPool.ReleaseBuffer(rawData)
	if rawData == nil || err != nil {
		return err
	}
	if err = s.conn.WriteAsync(*rawData); err != nil {
		return err
	}
	wsRequestCache.GetCache().StoreMeta("M"+api.ClientOrderId, &wsRequestCache.WsRequestMeta{
		Json:    string(*rawData),
		ReqType: wsRequestCache.MODIFY_ORDER,
		ReqFrom: reqFrom,
	})
	return err
}
