package orderSdkBnModel

import (
	"strconv"
	"time"

	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/infra/errorx/errDefine"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/utils/myCrypto"
	"upbitBnServer/pkg/container/pool/byteBufPool"
	"upbitBnServer/pkg/utils/convertx"
	"upbitBnServer/pkg/utils/timeUtils"

	"github.com/shopspring/decimal"
)

type FutureModifySdk struct {
	symbolName    string          //YES	交易对
	ClientOrderId string          //NO	用户自定义的订单号
	Quantity      decimal.Decimal //YES	下单数量,使用closePosition不支持此参数。
	Price         decimal.Decimal //YES	委托价格
	side          orderSide       //YES	买卖方向 SELL, BUY; side需要和原订单相同
}

func (api *FutureModifySdk) Symbol_(symbol string) *FutureModifySdk {
	api.symbolName = symbol
	return api
}

func (api *FutureModifySdk) ClientOrderId_(origClientOrderId string) *FutureModifySdk {
	api.ClientOrderId = origClientOrderId
	return api
}

func (api *FutureModifySdk) Side_(side orderSide) *FutureModifySdk {
	api.side = side
	return api
}

func (api *FutureModifySdk) Quantity_(quantity decimal.Decimal) *FutureModifySdk {
	api.Quantity = quantity
	return api
}

func (api *FutureModifySdk) Price_(price decimal.Decimal) *FutureModifySdk {
	api.Price = price
	return api
}

func (api *FutureModifySdk) ParseRestReqFast() *[]byte {
	orig := byteBufPool.AcquireBuffer(256)
	*orig = append(*orig, b_ORIG_CLIENT_ORDER_ID...)
	*orig = append(*orig, api.ClientOrderId...)

	*orig = append(*orig, b_SYMBOL...)
	*orig = append(*orig, api.symbolName...)

	*orig = append(*orig, b_SIDE...)
	*orig = append(*orig, orderSideArr[api.side]...)

	*orig = append(*orig, b_PRICE...)
	*orig = append(*orig, api.Price.String()...)

	*orig = append(*orig, b_QUANTITY...)
	*orig = append(*orig, api.Quantity.String()...)
	*orig = append(*orig, b_TIME_STAMP...)
	*orig = convertx.AppendValueToBytes(*orig, timeUtils.GetNowTimeUnixMilli())
	return orig
}

func (api *FutureModifySdk) ParseWsReqFast(apiKey string, secretByte []byte) (*[]byte, error) {
	if api.ClientOrderId == "" {
		return nil, errDefine.ClientOrderIdEmpty.WithMetadata(map[string]string{defineJson.ReqType: "FutureModifySdkWs"})
	}
	m := make(map[string]any)
	m[p_SYMBOL] = api.symbolName
	m[p_SIDE] = orderSideArr[api.side]
	m[p_PRICE] = api.Price.String()
	m[p_QUANTITY] = api.Quantity.String()
	m[p_ORIG_CLIENT_ORDER_ID] = api.ClientOrderId
	//统一逻辑
	m[p_API_KEY] = apiKey
	m[p_TIME_STAMP] = timeUtils.GetNowTimeUnixMilli()
	signRaw := buildQueryBytePool(256, m, modifySortedKeyFast) //从池子中获取256位签名数据
	signRes := byteBufPool.AcquireBuffer(64)                   //从池子中获取64位
	defer byteBufPool.ReleaseBuffer(signRaw)                   //释放签名数据
	defer byteBufPool.ReleaseBuffer(signRes)                   //释放签名值
	if err := myCrypto.HmacSha256Fast(secretByte, *signRaw, signRes); err != nil {
		return nil, err
	}
	return buildWsReqFast(512, "M"+api.ClientOrderId, "order.modify", m, modifySortedKeyFast, signRes), nil
}

var modifySortedKeyFast = []string{p_API_KEY, p_ORDER_ID, p_ORIG_CLIENT_ORDER_ID, p_PRICE, p_QUANTITY, p_SIDE, p_SYMBOL, p_TIME_STAMP}

func (api *FutureModifySdk) ParseWsReqFastNoSign() ([]byte, error) {
	if api.ClientOrderId == "" {
		return nil, errDefine.ClientOrderIdEmpty.WithMetadata(map[string]string{defineJson.ReqType: "FutureModifySdkWs"})
	}
	buf := make([]byte, 0, 512)
	buf = append(buf, `{"id":"M`...)
	buf = append(buf, api.ClientOrderId...)
	buf = append(buf, `","method":"order.modify","params":{"origClientOrderId":"`...)
	buf = append(buf, api.ClientOrderId...)
	buf = append(buf, `","price":"`...)
	buf = append(buf, api.Price.String()...)
	buf = append(buf, `","quantity":"`...)
	buf = append(buf, api.Quantity.String()...)
	buf = append(buf, `","side":"`...)
	buf = append(buf, orderSideArr[api.side]...)
	buf = append(buf, `","symbol":"`...)
	buf = append(buf, api.symbolName...)
	buf = append(buf, `","timestamp":"`...)
	buf = strconv.AppendInt(buf, time.Now().UnixMilli(), 10)
	buf = append(buf, `"}}`...)
	return buf, nil
}

// {
// 	"id": "Mtest123456",
// 	"method": "order.modify",
// 	"params": {
// 		"origClientOrderId": "test123456",
// 		"price": "4100",
// 		"quantity": "0.01",
// 		"side": "0",
// 		"symbol": "ETHUSDT",
// 		"timestamp": "1760274958009"
// 	}
// }

// NewFutureModifySdk  rest修改订单 (TRADE)
func NewFutureModifySdk() *FutureModifySdk {
	return &FutureModifySdk{}
}

func GetFutureModifySdk(req *orderModel.MyModifyOrderReq) *FutureModifySdk {
	side, _ := getBnOrderMode(req.OrderMode)
	api := NewFutureModifySdk().Symbol_(req.StaticMeta.SymbolName).Side_(side).Price_(req.ModifyPrice).Quantity_(req.OrigVol)
	if req.ClientOrderId != "" {
		api.ClientOrderId_(req.ClientOrderId)
	}
	return api
}
