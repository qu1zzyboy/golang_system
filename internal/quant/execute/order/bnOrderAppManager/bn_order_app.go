package bnOrderAppManager

import (
	"context"

	"github.com/hhh500/quantGoInfra/infra/observe/log/dynamicLog"
	"github.com/hhh500/quantGoInfra/infra/observe/notify"
	"github.com/hhh500/upbitBnServer/internal/quant/account/accountConfig"
	"github.com/hhh500/upbitBnServer/internal/quant/execute"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderSdk/bn/orderSdkBnRest"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderSdk/bn/orderSdkBnWsSign"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderStatic"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/wsRequestCache"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

type OnOrderPrePlace func(ok bool, clientOrderId string)
type OnMonitorData func(symbolIndex int, data []byte)
type OnFailureOrder func(accountKeyId uint8, errCode int64)
type OnCanceledOrder func(accountKeyId uint8)
type OnMaxWithdrawAmount func(accountKeyId uint8, maxWithdrawAmount decimal.Decimal)

var (
	fnPrePlace    OnOrderPrePlace                     // 预挂单回调函数
	fnSuccess     toUpBitListDataAfter.OnSuccessOrder // 下单成功回调函数
	fnMonitor     OnMonitorData                       // 监控数据回调函数
	fnFailure     OnFailureOrder                      // 下单失败回调函数
	fnCanceled    OnCanceledOrder                     // 撤单返回
	fnMaxWithdraw OnMaxWithdrawAmount                 //最大可划转金额返回
)

type OrderApp struct {
	rest         *orderSdkBnRest.FutureRest     // REST API 客户端
	wsOrderSign  *orderSdkBnWsSign.FutureClient // WS API 客户端
	accountKeyId uint8                          // 账户序号
	isMonitor    bool                           // 是否是监控账户
}

func newOrderApp() *OrderApp {
	return &OrderApp{}
}

func (s *OrderApp) init(ctx context.Context, v accountConfig.Config) error {
	s.accountKeyId = v.AccountId
	s.rest = orderSdkBnRest.NewFutureRest(v.ApiKeyHmac, v.SecretHmac)
	s.wsOrderSign = orderSdkBnWsSign.NewFutureClient(v.ApiKeyHmac, v.SecretHmac)
	if err := s.wsOrderSign.RegisterReadHandler(ctx, v.AccountId, s.OnWsOrder); err != nil {
		return err
	}
	return nil
}

func (s *OrderApp) OnWsOrder(data []byte) {
	idStr := gjson.GetBytes(data, "id").String()
	wsMeta, ok := wsRequestCache.GetCache().GetMeta(idStr)
	if !ok {
		// 帐号认证
		if gjson.GetBytes(data, "status").Int() == 200 {
			return
		}
		dynamicLog.Error.GetLog().Errorf("[%d]WS_REQUEST: [%s] req_id not found %s", s.accountKeyId, idStr, string(data))
		return
	}
	// 最终删除ws_meta
	defer wsRequestCache.GetCache().DelMeta(idStr)

	switch wsMeta.ReqType {
	case wsRequestCache.PLACE_ORDER:

		// 拿到clientOrderId去查内存静态数据
		clientOrderId := idStr[1:]
		orderFrom, orderMode, symbolIndex, ok := orderStatic.GetService().GetOrderInstanceIdAndSymbolId(clientOrderId)

		// 不属于这个服务的订单直接pass
		if !ok {
			// 账户登录的返回
			if gjson.GetBytes(data, "result.apiKey").Exists() {
				return
			}
			dynamicLog.Error.GetLog().Errorf("[%d]下单失败: [%s] orderFrom not found %s", s.accountKeyId, clientOrderId, string(data))
			return
		}

		// 请求是否成功
		ok = gjson.GetBytes(data, "status").Int() == 200

		switch orderFrom {
		case orderBelongEnum.TO_UPBIT_LIST_PRE:
			{
				if ok {
					isOnline := execute.IsOrderOnLine(execute.ParseBnOrderStatus(gjson.GetBytes(data, "result.status").String()))
					fnPrePlace(isOnline, clientOrderId)
				} else {
					code := gjson.GetBytes(data, "error.code").Int()
					// "code":-2013,"msg":"Order does not exist."
					// 预挂单失败,并且是订单不存在,直接忽略
					if code == (-2013) {
						return
					}
					// "code":-4116,"msg":"ClientOrderId is duplicated."
					// if code == (-4116) {
					// 	obj := toUpBitListDataBefore.GetSymbolDataObj(symbolIndex)
					// 	GetTradeManager().SendCancelOrder(s.accountKeyId, &orderModel.MyQueryOrderReq{
					// 		StaticMeta:    obj.StMeta,
					// 		ClientOrderId: clientOrderId,
					// 	})
					// }
					dynamicLog.Error.GetLog().Errorf("[%d]下单失败: 请求:%s,返回%s", s.accountKeyId, wsMeta.Json, string(data))
				}
			}
		case orderBelongEnum.TO_UPBIT_LIST_LOOP:
			{
				// 需要判断是不是触发的币种
				if symbolIndex != toUpBitListDataAfter.TrigSymbolIndex {
					/*
						一般都是因为网络太慢了,推送过来太慢了
					*/
					code := gjson.GetBytes(data, "error.code").Int()
					if toUpBitListDataAfter.TrigSymbolIndex == (-1) {
						// "code":-2027,"msg":"Exceeded the maximum allowable position at current leverage."
						if code == (-2027) {
							return
						}
						// "code":-5022,"msg":"Due to the order could not be executed as maker
						if code == (-5022) {
							return
						}
						// "code":-2019,"msg":"Margin is insufficient."
						if code == (-2019) {
							return
						}
						// "code":-4016,"msg":"Limit price can't be higher than 0.0012859."
						if code == (-4016) {
							return
						}
					}
					toUpBitListDataStatic.DyLog.GetLog().Errorf("[%d]触发后异常订单,请求数据:%s,返回数据%s", s.accountKeyId, wsMeta.Json, string(data))
					return
				}
				if ok {
					// 下单成功回调
					fnSuccess(toUpBitListDataAfter.OnSuccessEvt{
						ClientOrderId: clientOrderId,
						IsOnline:      true,
						OrderMode:     orderMode,
						InstanceId:    orderFrom,
						AccountKeyId:  s.accountKeyId,
						TimeStamp:     gjson.GetBytes(data, "result.updateTime").Int(),
					})
					return
				}
				errCode := gjson.GetBytes(data, "error.code").Int()
				// "Limit price can't be higher than 4550.62."
				// 价格超出下单失败
				if errCode == (-4016) {
					fnMonitor(symbolIndex, data)
					return
				}
				fnFailure(s.accountKeyId, errCode)
				// 如果已经触发,并且是只做maker失败,直接忽略
				// GTX_ORDER_REJECT
				if errCode == (-5022) {
					return
				}
				// "code":-2019,"msg":"Margin is insufficient."
				if errCode == (-2019) {
					return
				}
				toUpBitListDataStatic.DyLog.GetLog().Errorf("账户[%d]下单失败,请求:%s,返回:%s", s.accountKeyId, wsMeta.Json, string(data))
				notify.GetNotify().SendImportantErrorMsg(map[string]string{
					"msg":           "下单失败",
					"clientOrderId": clientOrderId,
				})
			}
		default:
			dynamicLog.Error.GetLog().Errorf("WS_REQUEST: unknown orderFrom %v", orderFrom)
		}

	case wsRequestCache.CANCEL_ORDER:
		switch wsMeta.ReqFrom {
		case orderBelongEnum.TO_UPBIT_LIST_LOOP_CANCEL_TRANSFER:
			{
				fnCanceled(s.accountKeyId)
			}
		case orderBelongEnum.TO_UPBIT_LIST_LOOP, orderBelongEnum.TO_UPBIT_LIST_PRE:
			{
				// 预期之内的行为
				return
			}
		default:
			dynamicLog.Error.GetLog().Errorf("CANCEL_ORDER: unknown orderFrom %v", wsMeta.ReqFrom)
		}

	case wsRequestCache.MODIFY_ORDER:
		{
			switch wsMeta.ReqFrom {
			case orderBelongEnum.TO_UPBIT_LIST_PRE:
				{
					// 预期之内的行为
					return
				}
			default:
				dynamicLog.Error.GetLog().Errorf("MODIFY_ORDER: unknown orderFrom %v", wsMeta.ReqFrom)
			}
		}
	case wsRequestCache.QUERY_ACCOUNT_BALANCE:
		{
			switch wsMeta.ReqFrom {
			case orderBelongEnum.TO_UPBIT_LIST_LOOP_CANCEL_TRANSFER:
				{
					value := gjson.GetBytes(data, `result.#(asset=="USDT").maxWithdrawAmount`)
					if value.Exists() {
						fnMaxWithdraw(s.accountKeyId, decimal.RequireFromString(value.String()))
					} else {
						dynamicLog.Error.GetLog().Errorf("QUERY_ACCOUNT_BALANCE: json异常 %s", string(data))
					}
				}
			default:
				dynamicLog.Error.GetLog().Errorf("QUERY_ACCOUNT_BALANCE: unknown orderFrom %v", wsMeta.ReqFrom)
			}
		}
	default:
		dynamicLog.Error.GetLog().Errorf("WS_REQUEST: unknown WsRequestType %v", wsMeta.ReqType)
	}
}

// {
// 	"id": "test123456",
// 	"status": 400,
// 	"error": {
// 		"code": -4003,
// 		"msg": "Quantity less than or equal to zero."
// 	},
// 	"rateLimits": [{
// 		"rateLimitType": "REQUEST_WEIGHT",
// 		"interval": "MINUTE",
// 		"intervalNum": 1,
// 		"limit": -1,
// 		"count": -1
// 	}, {
// 		"rateLimitType": "ORDERS",
// 		"interval": "SECOND",
// 		"intervalNum": 10,
// 		"limit": 300,
// 		"count": 1
// 	}, {
// 		"rateLimitType": "ORDERS",
// 		"interval": "MINUTE",
// 		"intervalNum": 1,
// 		"limit": 1200,
// 		"count": 1
// 	}]
// }

// {
// 	"id": "test123456",
// 	"status": 200,
// 	"result": {
// 		"orderId": 8222959323,
// 		"symbol": "ETHUSDC",
// 		"status": "NEW",
// 		"clientOrderId": "test123456",
// 		"price": "1800.00",
// 		"avgPrice": "0.00",
// 		"origQty": "1.000",
// 		"executedQty": "0.000",
// 		"cumQuote": "0.00000",
// 		"timeInForce": "GTC",
// 		"type": "LIMIT",
// 		"reduceOnly": false,
// 		"closePosition": false,
// 		"side": "BUY",
// 		"positionSide": "LONG",
// 		"stopPrice": "0.00",
// 		"workingType": "CONTRACT_PRICE",
// 		"priceProtect": false,
// 		"origType": "LIMIT",
// 		"priceMatch": "NONE",
// 		"selfTradePreventionMode": "EXPIRE_MAKER",
// 		"goodTillDate": 0,
// 		"time": 1743216918241,
// 		"updateTime": 1743216918241
// 	},
// 	"rateLimits": [{
// 		"rateLimitType": "REQUEST_WEIGHT",
// 		"interval": "MINUTE",
// 		"intervalNum": 1,
// 		"limit": 2400,
// 		"count": 6
// 	}]
// }
