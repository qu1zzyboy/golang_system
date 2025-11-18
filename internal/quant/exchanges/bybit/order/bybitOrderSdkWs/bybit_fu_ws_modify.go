package bybitOrderSdkWs

import (
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/quant/exchanges/bybit/order/bybitQueryOrder"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/quant/execute/order/wsRequestCache"
	"upbitBnServer/pkg/utils/time2str"
)

func (s *FutureClient) ModifyOrder(req orderModel.MyModifyOrderReq) error {
	var rawData []byte
	reqId := time2str.GetNowTimeStampMicroSlice16()
	rawData = bybitQueryOrder.GetByBitFu_M(req.SymbolName, req.Pvalue, req.Qvalue, req.Pscale, req.Qscale, reqId, req.ClientOrderId)
	if err := s.conn.WriteAsync(rawData); err != nil {
		dynamicLog.Error.GetLog().Errorf("bybit_ws改单错误,请求:%s,错误:%v", string(rawData), err)
		return err
	}
	reqId[0] = 'M'
	wsRequestCache.GetCache().StoreMeta(reqId, wsRequestCache.WsRequestMeta{
		ReqJson:   rawData,
		ReqType:   wsRequestCache.MODIFY_ORDER,
		ReqFrom:   req.ReqFrom,
		UsageFrom: req.UsageFrom,
	})
	return nil
}
