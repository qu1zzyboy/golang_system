package orderSdkBnWsSign

import (
	"bytes"
	"time"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/quant/execute/order/orderSdk/bn/orderSdkBnModel"
	"upbitBnServer/internal/quant/execute/order/orderStaticMeta"
	"upbitBnServer/internal/quant/execute/order/wsRequestCache"
	"upbitBnServer/pkg/container/pool/byteBufPool"
)

func (s *FutureClient) CreateOrder(req *orderModel.MyPlaceOrderReq) error {
	api := orderSdkBnModel.GetFuturePlaceLimitSdk(req)
	rawData, err := api.ParseWsReqFast(s.apiKey, s.secretByte)
	defer byteBufPool.ReleaseBuffer(rawData)
	if rawData == nil || err != nil {
		return err
	}
	if err = s.conn.WriteAsync(*rawData); err != nil {
		return err
	}
	orderStaticMeta.GetService().SaveOrderMeta(req.ClientOrderId, orderStaticMeta.StaticMeta{
		SymbolIndex:  req.SymbolIndex,
		OrderMode:    req.OrderMode,
		InstanceFrom: req.ReqFrom,
		UsageFrom:    req.UsageFrom,
	})
	wsRequestCache.GetCache().StoreMeta(systemx.WsId16B("P"+api.ClientOrderId), &wsRequestCache.WsRequestMeta{
		ReqJson:       bytes.Clone(*rawData),
		ClientOrderId: req.ClientOrderId,
		UpdateAt:      time.Now().UnixMilli(),
		ReqType:       wsRequestCache.PLACE_ORDER,
		ReqFrom:       req.ReqFrom,
		UsageFrom:     req.UsageFrom,
	})
	return nil
}
