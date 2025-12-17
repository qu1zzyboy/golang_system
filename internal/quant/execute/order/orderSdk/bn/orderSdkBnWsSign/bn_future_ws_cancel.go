package orderSdkBnWsSign

import (
	"bytes"
	"time"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/quant/execute/order/orderSdk/bn/orderSdkBnModel"
	"upbitBnServer/internal/quant/execute/order/wsRequestCache"
	"upbitBnServer/pkg/container/pool/byteBufPool"
)

func (s *FutureClient) CancelOrder(req *orderModel.MyQueryOrderReq) error {
	api := orderSdkBnModel.GetFutureQuerySdk(req)
	rawData, err := api.ParseWsReqFast("C", s.apiKey, "order.cancel", s.secretByte)
	defer byteBufPool.ReleaseBuffer(rawData)
	if rawData == nil || err != nil {
		return err
	}
	if err = s.conn.WriteAsync(*rawData); err != nil {
		return err
	}
	wsRequestCache.GetCache().StoreMeta(systemx.WsId16B("C"+api.ClientOrderId), &wsRequestCache.WsRequestMeta{
		ReqJson:       bytes.Clone(*rawData),
		ClientOrderId: req.ClientOrderId,
		UpdateAt:      time.Now().UnixMilli(),
		ReqType:       wsRequestCache.CANCEL_ORDER,
		ReqFrom:       req.ReqFrom,
		UsageFrom:     req.UsageFrom,
	})
	return err
}
