package orderSdkBybitWs

import (
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/quant/execute/order/orderSdk/bybit/queryOrderByBit"
	"upbitBnServer/internal/quant/execute/order/wsRequestCache"
	"upbitBnServer/pkg/utils/time2str"
)

func (s *FutureClient) CancelOrder(req orderModel.MyQueryOrderReq) error {
	reqId := time2str.GetNowTimeStampMicroSlice16()
	rawData := queryOrderByBit.GetByBitFu_C(req.SymbolName, reqId, req.ClientOrderId)
	if err := s.conn.WriteAsync(rawData); err != nil {
		dynamicLog.Error.GetLog().Errorf("bybit_ws撤单错误,请求:%s,错误:%v", string(rawData), err)
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
