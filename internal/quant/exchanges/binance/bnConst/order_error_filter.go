package bnConst

var (
	orderErrCode = map[string]struct{}{
		"-2018": {}, // BALANCE_NOT_SUFFICIENT
		"-2019": {}, // MARGIN_NOT_SUFFICIENT
		"-4050": {}, // CROSS_BALANCE_INSUFFICIENT
		"-4051": {}, // ISOLATED_BALANCE_INSUFFICIENT
		"-2022": {}, // REDUCE_ONLY_REJECT
		"-4118": {}, // REDUCE_ONLY_MARGIN_CHECK_FAILED
		"-5022": {}, // GTX_ORDER_REJECT
		"-5021": {}, // FOK_ORDER_REJECT
		"-2021": {}, // ORDER_WOULD_IMMEDIATELY_TRIGGER
	}
)

// -1001 DISCONNECTED Internal error; unable to process your request. Please try again.
// -5028 ME_RECVWINDOW_REJECT 请求的时间戳在撮合的recvWindow之外
//-1008 Server is currently overloaded with other requests. Please try again in a few minutes.
//-1015 Too many new orders; current limit is 300
// 下单
//改单
//-2013 Order does not exist (订单不存在)
//-1008 Server is currently overloaded with other requests. Please try again in a few minutes.
//-1015 Too many new orders; current limit is 300
// {"code":-5027,"msg":"No need to modify the order."}

func IsOrderErrCodeFilter(code string) bool {
	_, ok := orderErrCode[code]
	return ok
}
