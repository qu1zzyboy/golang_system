package bnOrderSdkModel

import (
	"fmt"
	"net/url"
	"strings"

	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/pkg/utils/convertx"
	"upbitBnServer/pkg/utils/jsonUtils"
)

type FutureCancelBatchSdk struct {
	symbolName        string   //YES 交易对
	orderIdList       []int64  //NO 系统订单号, 最多支持10个订单
	ClientOrderIdList []string //NO 用户自定义的订单号, 最多支持10个订单
}

func (api *FutureCancelBatchSdk) Symbol_(Symbol string) *FutureCancelBatchSdk {
	api.symbolName = Symbol
	return api
}

func (api *FutureCancelBatchSdk) OrderIdList_(orderIDList []int64) *FutureCancelBatchSdk {
	api.orderIdList = orderIDList
	return api
}

func (api *FutureCancelBatchSdk) ClientOrderIdList_(origClientOrderIdList []string) *FutureCancelBatchSdk {
	api.ClientOrderIdList = origClientOrderIdList
	return api
}

func (api *FutureCancelBatchSdk) ParseRestRequest() string {
	param := url.Values{}
	param.Set(p_SYMBOL, api.symbolName)
	if api.orderIdList != nil {
		// convert a slice of integers to a string e.g. [1 2 3] => "[1,2,3]"
		orderIDListString := strings.Join(strings.Fields(fmt.Sprint(api.orderIdList)), ",")
		param.Add(p_ORDER_ID_LIST, orderIDListString) //原来是r.form表单提交
	}
	if api.ClientOrderIdList != nil {
		data, _ := jsonUtils.MarshalStructToByteArray(api.ClientOrderIdList)
		// data, _ := json.Marshal(api.req.ClientOrderIdList_)
		// 去掉 JSON 生成的多余空格(一般不会有)
		param.Add(p_ORIG_CLIENT_ORDER_ID_LIST, strings.ReplaceAll(string(data), " ", ""))
	}
	param.Set(p_TIME_STAMP, convertx.GetNowTimeStampMilliStr())
	// 编码 query & form
	return param.Encode()
}

// NewFutureCancelBatchSdk  rest批量撤销订单 (TRADE)
func NewFutureCancelBatchSdk() *FutureCancelBatchSdk {
	return &FutureCancelBatchSdk{}
}

func GetFutureCancelBatchSdk(req *orderModel.MyCancelOrderBatchReq) *FutureCancelBatchSdk {
	return NewFutureCancelBatchSdk().Symbol_(req.StaticMeta.SymbolName).ClientOrderIdList_(req.GetClientOrderIds())
}
