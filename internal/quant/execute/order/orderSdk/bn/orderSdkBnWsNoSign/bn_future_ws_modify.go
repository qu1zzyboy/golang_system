package orderSdkBnWsNoSign

import (
	"bytes"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/systemx/instanceEnum"
	"upbitBnServer/internal/infra/systemx/usageEnum"
	"upbitBnServer/internal/quant/execute/order/bnOrderTemplate"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/quant/execute/order/orderSdk/modifyBuyLimit"
	"upbitBnServer/internal/quant/execute/order/orderSdk/modifySellLimit"
	"upbitBnServer/internal/quant/execute/order/wsRequestCache"
	"upbitBnServer/pkg/utils/time2str"
)

func (s *FutureClient) ModifyOrder(req orderModel.MyModifyOrderReq) error {
	var rawData []byte
	reqId := time2str.GetNowTimeStampMicroSlice16()
	if req.OrderMode.IsBuy() {
		rawData = modifyBuyLimit.GetBnFu_NoSign_u32(req.SymbolName, req.Pvalue, req.Qvalue, req.Pscale, req.Qscale, reqId, req.ClientOrderId)
	} else {
		rawData = modifySellLimit.GetBnFu_NoSign_u32(req.SymbolName, req.Pvalue, req.Qvalue, req.Pscale, req.Qscale, reqId, req.ClientOrderId)
	}
	if err := s.conn.WriteAsync(rawData); err != nil {
		dynamicLog.Error.GetLog().Errorf("bn_ws改单错误,请求:%s,错误:%v", string(rawData), err)
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

func (s *FutureClient) ModifyOrderBy(modify *bnOrderTemplate.ModifyTemplate, reqFrom instanceEnum.Type, usageFrom usageEnum.Type) error {
	reqId := time2str.GetNowTimeStampMicroSlice16()
	rawData := modify.GetModifyRaw(reqId)
	if err := s.conn.WriteAsync(rawData); err != nil {
		dynamicLog.Error.GetLog().Errorf("bn_ws改单错误,请求:%s,错误:%v", string(rawData), err)
		return err
	}
	reqId[0] = 'M'
	wsRequestCache.GetCache().StoreMeta(reqId, wsRequestCache.WsRequestMeta{
		ReqJson:   bytes.Clone(rawData),
		ReqType:   wsRequestCache.MODIFY_ORDER,
		ReqFrom:   reqFrom,
		UsageFrom: usageFrom,
	})
	return nil
}
