package bnOrderSdkWsNoSign

import (
	"bytes"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/systemx/instanceEnum"
	"upbitBnServer/internal/infra/systemx/usageEnum"
	"upbitBnServer/internal/quant/exchanges/binance/order/bnOrderTemplate"
	"upbitBnServer/internal/quant/exchanges/binance/order/bnQueryOrder"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/quant/execute/order/wsRequestCache"
	"upbitBnServer/pkg/utils/time2str"
)

func (s *FutureClient) QueryOrder(req orderModel.MyQueryOrderReq) error {
	reqId := time2str.GetNowTimeStampMicroSlice16()
	rawData := bnQueryOrder.GetBnFu_NoSign_Q(req.SymbolName, reqId, req.ClientOrderId)
	if err := s.conn.WriteAsync(rawData); err != nil {
		dynamicLog.Error.GetLog().Errorf("bn_ws查单错误,请求:%s,错误:%v", string(rawData), err)
		return err
	}
	reqId[0] = 'Q'
	wsRequestCache.GetCache().StoreMeta(reqId, wsRequestCache.WsRequestMeta{
		ReqJson:   bytes.Clone(rawData),
		ReqType:   wsRequestCache.QUERY_ORDER,
		ReqFrom:   req.ReqFrom,
		UsageFrom: req.UsageFrom,
	})
	return nil
}

func (s *FutureClient) CancelOrder(req orderModel.MyQueryOrderReq) error {
	reqId := time2str.GetNowTimeStampMicroSlice16()
	rawData := bnQueryOrder.GetBnFu_NoSign_C(req.SymbolName, reqId, req.ClientOrderId)
	if err := s.conn.WriteAsync(rawData); err != nil {
		dynamicLog.Error.GetLog().Errorf("bn_ws撤单错误,请求:%s,错误:%v", string(rawData), err)
		return err
	}
	reqId[0] = 'C'
	wsRequestCache.GetCache().StoreMeta(reqId, wsRequestCache.WsRequestMeta{
		ReqJson:   rawData,
		ReqType:   wsRequestCache.CANCEL_ORDER,
		ReqFrom:   req.ReqFrom,
		UsageFrom: req.UsageFrom,
	})
	return nil
}

func (s *FutureClient) CancelOrderBy(can *bnOrderTemplate.CancelTemplate, reqFrom instanceEnum.Type, usageFrom usageEnum.Type) error {
	reqId := time2str.GetNowTimeStampMicroSlice16()
	rawData := can.GetCancelRaw(reqId)
	if err := s.conn.WriteAsync(rawData); err != nil {
		dynamicLog.Error.GetLog().Errorf("bn_ws撤单错误,请求:%s,错误:%v", string(rawData), err)
		return err
	}
	reqId[0] = 'C'
	wsRequestCache.GetCache().StoreMeta(reqId, wsRequestCache.WsRequestMeta{
		ReqJson:   bytes.Clone(rawData),
		ReqType:   wsRequestCache.CANCEL_ORDER,
		ReqFrom:   reqFrom,
		UsageFrom: usageFrom,
	})
	return nil
}
