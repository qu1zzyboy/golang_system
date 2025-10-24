package orderSdkBnModel

import (
	"upbitBnServer/pkg/container/pool/byteBufPool"
	"upbitBnServer/pkg/utils/convertx"
)

const (
	keyType = "HMAC"
)

const (
	p_ID                  = "id"
	p_API_KEY             = "apiKey"
	p_TIME_STAMP          = "timestamp"
	p_SYMBOL              = "symbol"
	p_TYPE                = "type"
	p_SIDE                = "side"
	p_NEW_CLIENT_ORDER_ID = "newClientOrderId"
	p_TIME_IN_FORCE       = "timeInForce"
	p_PRICE               = "price"
	p_STOP_PRICE          = "stopPrice"
	p_QUANTITY            = "quantity"
	p_NEW_ORDER_RESP_TYPE = "newOrderRespType"
	p_POSITION_SIDE       = "positionSide"
)

var (
	b_ID_ = []byte(`{"id":"`)

	b_METHOD_ = []byte(`","method":"`)
	b_PARAMS_ = []byte(`","toUpbitParam":{`)
	b_WS_STEP = []byte(`":"`)
	b_WS_END  = []byte("}}")
)

const (
	p_SIGNATURE_KEY             = "signature"
	p_RECEIVE_WINDOW            = "receiveWindow"
	p_ORDER_ID                  = "orderId"
	p_ORIG_CLIENT_ORDER_ID      = "origClientOrderId"
	p_ORIG_CLIENT_ORDER_ID_LIST = "origClientOrderIdList"
	p_ORDER_ID_LIST             = "orderIdList"
	p_BATCH_ORDERS              = "batchOrders"
)

var (
	b_API_KEY                   = []byte("apiKey=")
	b_SIGNATURE                 = []byte("signature=")
	b_RECEIVE_WINDOW            = []byte("&receiveWindow=")
	b_ORIG_CLIENT_ORDER_ID_LIST = []byte("origClientOrderIdList=")
	b_ORDER_ID_LIST_Byte        = []byte("orderIdList=")
	b_BATCH_ORDERS_Byte         = []byte("batchOrders=")
	b_PLACE_WS                  = []byte("order.place")
	b_MODIFY_WS                 = []byte("order.modify")
)

func buildWsReqFast(preSize int, wsRequestId, method string, params map[string]any, keySorted []string, signData *[]byte) *[]byte {
	b := byteBufPool.AcquireBuffer(preSize)
	*b = append(*b, b_ID_...) //{"id":"
	*b = append(*b, wsRequestId...)
	*b = append(*b, b_METHOD_...) //","method":"
	*b = append(*b, method...)
	*b = append(*b, b_PARAMS_...) //","toUpbitParam":{
	for i, k := range keySorted {
		if val, ok := params[k]; ok {
			if i > 0 {
				*b = append(*b, ',')
			}
			*b = append(*b, '"')
			*b = append(*b, k...)
			*b = append(*b, b_WS_STEP...)
			*b = convertx.AppendValueToBytes(*b, val)
			*b = append(*b, '"')
		}
	}
	// 添加签名字段
	*b = append(*b, ',')
	*b = append(*b, `"signature":"`...)
	*b = append(*b, *signData...)
	*b = append(*b, '"')
	*b = append(*b, b_WS_END...)
	return b
}

func buildWsReqFastNoSign(preSize int, wsRequestId, method string, params map[string]any, keySorted []string) *[]byte {
	b := byteBufPool.AcquireBuffer(preSize)
	*b = append(*b, b_ID_...) //{"id":"
	*b = append(*b, wsRequestId...)
	*b = append(*b, b_METHOD_...) //","method":"
	*b = append(*b, method...)
	*b = append(*b, b_PARAMS_...) //","toUpbitParam":{
	for i, k := range keySorted {
		if val, ok := params[k]; ok {
			if i > 0 {
				*b = append(*b, ',')
			}
			*b = append(*b, '"')
			*b = append(*b, k...)
			*b = append(*b, b_WS_STEP...)
			*b = convertx.AppendValueToBytes(*b, val)
			*b = append(*b, '"')
		}
	}
	*b = append(*b, b_WS_END...)
	return b
}

func buildQueryBytePool(preSize int, params map[string]any, keySorted []string) *[]byte {
	b := byteBufPool.AcquireBuffer(preSize)
	for i, k := range keySorted {
		if val, ok := params[k]; ok {
			if i > 0 {
				*b = append(*b, '&')
			}
			*b = append(*b, k...)
			*b = append(*b, '=')
			*b = convertx.AppendValueToBytes(*b, val)
		}
	}
	return b
}
