package toUpbitByBitWsParse

import (
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/observe/notify"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/usageEnum"
	"upbitBnServer/internal/quant/execute/order/orderStatic"
	"upbitBnServer/internal/quant/execute/order/wsRequestCache"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListChan"

	"github.com/shopspring/decimal"
)

func (s *Parser) onPlaceOrderFailed(data []byte, wsMeta wsRequestCache.WsRequestMeta, clientOrderId systemx.WsId16B, accountKeyId uint8) {
	oMeta, ok := orderStatic.GetService().GetOrderMeta(clientOrderId)
	// 不属于这个服务的订单直接pass
	if !ok {
		dynamicLog.Error.GetLog().Errorf("[%d]PLACE_ORDER: [%s] orderFrom not found %s", accountKeyId, string(clientOrderId[:]), string(data))
		return
	}
	symbolIndex := oMeta.SymbolIndex
	usage := wsMeta.UsageFrom

	switch usage {
	case usageEnum.TO_UPBIT_PRE:
		{
			dynamicLog.Error.GetLog().Errorf("[%d]下单失败: 请求:%s,返回%s", accountKeyId, string(wsMeta.ReqJson), string(data))
		}
	case usageEnum.TO_UPBIT_MAIN:
		{
			// 需要判断是不是触发的币种
			if symbolIndex != toUpBitListDataAfter.TrigSymbolIndex {
				/*
					一般都是因为网络太慢了,推送过来太慢了
				*/
				toUpBitDataStatic.DyLog.GetLog().Errorf("[%d]触发后异常订单,请求数据:%s,返回数据%s", accountKeyId, string(wsMeta.ReqJson), string(data))
				return
			}

			// 发送错误码
			toUpbitListChan.SendSpecial(symbolIndex, toUpbitListChan.Special{
				Amount:       decimal.Zero,
				ErrCode:      [5]byte{data[55], data[56], data[57], data[58]},
				SigType:      toUpbitListChan.FailureOrder,
				AccountKeyId: accountKeyId,
			})
			/**********过滤掉常见错误，然后打印************/
			toUpBitDataStatic.DyLog.GetLog().Errorf("账户[%d]下单失败,请求:%s,返回:%s", accountKeyId, string(wsMeta.ReqJson), string(data))
			notify.GetNotify().SendImportantErrorMsg(map[string]string{"msg": "下单失败", "clientOrderId": string(clientOrderId[:])})
		}
	default:
		dynamicLog.Error.GetLog().Errorf("PLACE_ORDER: unknown orderFrom %v", oMeta.UsageFrom)
	}
}

// 挂单成功
// {
//   "id": "1761662304009422",
//   "status": 200,
//   "result": {
//     "orderId": 8389765998331620000,
//     "symbol": "ETHUSDT",
//     "status": "NEW",
//     "clientOrderId": "1761662304009422",
//     "price": "3100.00",
//     "avgPrice": "0.00",
//     "origQty": "0.010",
//     "executedQty": "0.000",
//     "cumQty": "0.000",
//     "cumQuote": "0.00000",
//     "timeInForce": "GTC",
//     "type": "LIMIT",
//     "reduceOnly": false,
//     "closePosition": false,
//     "side": "BUY",
//     "positionSide": "LONG",
//     "stopPrice": "0.00",
//     "workingType": "CONTRACT_PRICE",
//     "priceProtect": false,
//     "origType": "LIMIT",
//     "priceMatch": "NONE",
//     "selfTradePreventionMode": "EXPIRE_MAKER",
//     "goodTillDate": 0,
//     "updateTime": 1761662304015
//   }
// }

// 探测返回
// {
//   "id": "1761660897009235",
//   "status": 400,
//   "error": {
//     "code": -4016,
//     "msg": "Limit price can't be higher than 4370.70."
//   }
// }

// {"id":"Pfu2-A-EVAA7389116673233420540","status":400,"error":{"code":-4014,"msg":"Price not increased by tick size."}}
// {"id":"Pfu8-l-maker7388831890662129920","status":400,"error":{"code":-2027,"msg":"Exceeded the maximum allowable position at current leverage."}}
// {"id":"Pfu3-D-TRUTH7389116673233420542","status":400,"error":{"code":-4400,"msg":"Futures Trading Quantitative Rules violated, only reduceOnly order is allowed, please try again later."}}
