package orderSdkBnWsNoSign

import (
	"bytes"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/quant/execute/order/orderSdk/buyCloseLimit"
	"upbitBnServer/internal/quant/execute/order/orderSdk/buyCloseLimitMaker"
	"upbitBnServer/internal/quant/execute/order/orderSdk/buyOpenLimit"
	"upbitBnServer/internal/quant/execute/order/orderSdk/buyOpenLimitMaker"
	"upbitBnServer/internal/quant/execute/order/orderSdk/buyOpenMarket"
	"upbitBnServer/internal/quant/execute/order/orderSdk/sellCloseLimit"
	"upbitBnServer/internal/quant/execute/order/orderSdk/sellCloseLimitMaker"
	"upbitBnServer/internal/quant/execute/order/orderSdk/sellOpenLimit"
	"upbitBnServer/internal/quant/execute/order/orderSdk/sellOpenLimitMaker"
	"upbitBnServer/internal/quant/execute/order/orderSdk/sellOpenMarket"
	"upbitBnServer/internal/quant/execute/order/orderStatic"
	"upbitBnServer/internal/quant/execute/order/wsRequestCache"
)

func (s *FutureClient) CreateOrder(req orderModel.MyPlaceOrderReq) error {
	var rawData []byte
	switch req.OrderMode {
	case execute.BUY_OPEN_LIMIT:
		rawData = buyOpenLimit.GetBnFu_NoSign_u32(req.SymbolName, req.Pvalue, req.Qvalue, req.Pscale, req.Qscale, req.ClientOrderId)
	case execute.SELL_OPEN_LIMIT:
		rawData = sellOpenLimit.GetBnFu_NoSign_u32(req.SymbolName, req.Pvalue, req.Qvalue, req.Pscale, req.Qscale, req.ClientOrderId)
	case execute.BUY_CLOSE_LIMIT:
		rawData = buyCloseLimit.GetBnFu_NoSign_u32(req.SymbolName, req.Pvalue, req.Qvalue, req.Pscale, req.Qscale, req.ClientOrderId)
	case execute.SELL_CLOSE_LIMIT:
		rawData = sellCloseLimit.GetBnFu_NoSign_u32(req.SymbolName, req.Pvalue, req.Qvalue, req.Pscale, req.Qscale, req.ClientOrderId)
	case execute.BUY_OPEN_LIMIT_MAKER:
		rawData = buyOpenLimitMaker.GetBnFu_NoSign_u32(req.SymbolName, req.Pvalue, req.Qvalue, req.Pscale, req.Qscale, req.ClientOrderId)
	case execute.SELL_OPEN_LIMIT_MAKER:
		rawData = sellOpenLimitMaker.GetBnFu_NoSign_u32(req.SymbolName, req.Pvalue, req.Qvalue, req.Pscale, req.Qscale, req.ClientOrderId)
	case execute.BUY_CLOSE_LIMIT_MAKER:
		rawData = buyCloseLimitMaker.GetBnFu_NoSign_u32(req.SymbolName, req.Pvalue, req.Qvalue, req.Pscale, req.Qscale, req.ClientOrderId)
	case execute.SELL_CLOSE_LIMIT_MAKER:
		rawData = sellCloseLimitMaker.GetBnFu_NoSign_u32(req.SymbolName, req.Pvalue, req.Qvalue, req.Pscale, req.Qscale, req.ClientOrderId)
	case execute.BUY_OPEN_MARKET:
		rawData = buyOpenMarket.GetBnFu_NoSign_u32(req.SymbolName, req.Qvalue, req.Qscale, req.ClientOrderId)
	case execute.SELL_OPEN_MARKET:
		rawData = sellOpenMarket.GetBnFu_NoSign_u32(req.SymbolName, req.Qvalue, req.Qscale, req.ClientOrderId)
	}
	if err := s.conn.WriteAsync(rawData); err != nil {
		dynamicLog.Error.GetLog().Errorf("bn_ws下单错误,请求:%s,错误:%v", string(rawData), err)
		return err
	}
	orderStatic.GetService().SaveOrderMeta(req.ClientOrderId, orderStatic.StaticMeta{
		Pvalue:      req.Pvalue,
		Qvalue:      req.Qvalue,
		SymbolIndex: req.SymbolIndex,
		SymbolLen:   req.SymbolLen,
		OrderMode:   req.OrderMode,
		ReqFrom:     req.ReqFrom,
		UsageFrom:   req.UsageFrom,
	})
	wsRequestCache.GetCache().StoreMeta(req.ClientOrderId, wsRequestCache.WsRequestMeta{
		ReqJson:   bytes.Clone(rawData),
		ReqType:   wsRequestCache.PLACE_ORDER,
		ReqFrom:   req.ReqFrom,
		UsageFrom: req.UsageFrom,
	})
	return nil
}
