package orderSdkBnModel

import (
	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/infra/errorx/errDefine"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/utils/myCrypto"
	"upbitBnServer/pkg/container/pool/byteBufPool"
	"upbitBnServer/pkg/utils/convertx"
	"upbitBnServer/pkg/utils/timeUtils"
)

type FutureQuerySdk struct {
	symbolName    string //YES 交易对
	ClientOrderId string //NO 用户自定义的订单号
	orderId       *int64 //NO 系统订单号
}

func (api *FutureQuerySdk) Symbol_(symbol string) *FutureQuerySdk {
	api.symbolName = symbol
	return api
}

func (api *FutureQuerySdk) OrderId_(orderId int64) *FutureQuerySdk {
	api.orderId = &orderId
	return api
}

func (api *FutureQuerySdk) ClientOrderId_(clientOrderId string) *FutureQuerySdk {
	api.ClientOrderId = clientOrderId
	return api
}

// ParseRestReqFast 172.3 ns/op           184 B/op          5 allocs/op
func (api *FutureQuerySdk) ParseRestReqFast() *[]byte {
	orig := byteBufPool.AcquireBuffer(128)
	*orig = append(*orig, b_ORIG_CLIENT_ORDER_ID...)
	*orig = append(*orig, api.ClientOrderId...)

	*orig = append(*orig, b_SYMBOL...)
	*orig = append(*orig, api.symbolName...)
	if api.orderId != nil {
		*orig = append(*orig, b_ORDER_ID...)
		*orig = convertx.AppendValueToBytes(*orig, *api.orderId)
	}
	*orig = append(*orig, b_TIME_STAMP...)
	*orig = convertx.AppendValueToBytes(*orig, timeUtils.GetNowTimeUnixMilli())
	return orig
}

var querySortedKeyFast = []string{p_API_KEY, p_ORDER_ID, p_ORIG_CLIENT_ORDER_ID, p_SYMBOL, p_TIME_STAMP}

func (api *FutureQuerySdk) ParseWsReqFast(firstChar, apiKey, method string, secretByte []byte) (*[]byte, error) {
	if api.ClientOrderId == "" {
		return nil, errDefine.ClientOrderIdEmpty.WithMetadata(map[string]string{defineJson.ReqType: "FutureQuerySdkWs"})
	}
	param := make(map[string]any)
	param[p_SYMBOL] = api.symbolName
	param[p_ORIG_CLIENT_ORDER_ID] = api.ClientOrderId
	if api.orderId != nil {
		param[p_ORDER_ID] = *api.orderId
	}
	//统一逻辑
	param[p_API_KEY] = apiKey
	param[p_TIME_STAMP] = timeUtils.GetNowTimeUnixMilli()
	signRaw := buildQueryBytePool(128, param, querySortedKeyFast) //从池子中获取128位签名数据
	signRes := byteBufPool.AcquireBuffer(64)                      //从池子中获取64位
	defer byteBufPool.ReleaseBuffer(signRaw)                      //释放签名数据
	defer byteBufPool.ReleaseBuffer(signRes)                      //释放签名值
	if err := myCrypto.HmacSha256Fast(secretByte, *signRaw, signRes); err != nil {
		return nil, err
	}
	return buildWsReqFast(512, firstChar+api.ClientOrderId, method, param, querySortedKeyFast, signRes), nil
}

var querySortedKeyFastNoSign = []string{p_ORIG_CLIENT_ORDER_ID, p_SYMBOL, p_TIME_STAMP}

func (api *FutureQuerySdk) ParseWsReqFastNoSign(firstChar, method string) (*[]byte, error) {
	if api.ClientOrderId == "" {
		return nil, errDefine.ClientOrderIdEmpty.WithMetadata(map[string]string{defineJson.ReqType: "ParseWsReqFastNoSign"})
	}
	param := make(map[string]any)
	param[p_SYMBOL] = api.symbolName
	param[p_ORIG_CLIENT_ORDER_ID] = api.ClientOrderId
	if api.orderId != nil {
		param[p_ORDER_ID] = *api.orderId
	}
	param[p_TIME_STAMP] = timeUtils.GetNowTimeUnixMilli()
	return buildWsReqFastNoSign(512, firstChar+api.ClientOrderId, method, param, querySortedKeyFastNoSign), nil
}

// NewFutureQuerySdk   rest查询订单 (USER_DATA)
func NewFutureQuerySdk() *FutureQuerySdk {
	return &FutureQuerySdk{}
}

func GetFutureQuerySdk(req *orderModel.MyQueryOrderReq) *FutureQuerySdk {
	return NewFutureQuerySdk().Symbol_(req.StaticMeta.SymbolName).ClientOrderId_(req.ClientOrderId)
}
