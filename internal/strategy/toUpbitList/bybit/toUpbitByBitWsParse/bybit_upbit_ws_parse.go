package toUpbitByBitWsParse

import (
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/usageEnum"
	"upbitBnServer/internal/quant/execute/order/orderStatic"
	"upbitBnServer/internal/quant/execute/order/wsRequestCache"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListChan"
	"upbitBnServer/pkg/singleton"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

var serviceSingleton = singleton.NewSingleton(func() *Parser {
	return &Parser{}
})

func Get() *Parser {
	return serviceSingleton.Get()
}

type Parser struct {
}

func (s *Parser) Parse(data []byte, wsMeta wsRequestCache.WsRequestMeta, reqId systemx.WsId16B, accountKeyId uint8) {
	reqOk := data[38] == '0'
	usage := wsMeta.UsageFrom

	switch wsMeta.ReqType {
	case wsRequestCache.PLACE_ORDER:
		//下单成功由payload处理
		if reqOk {
			return
		}
		s.onPlaceOrderFailed(data, wsMeta, reqId, accountKeyId)

		// 删除所有下单失败的订单
		orderStatic.GetService().DelOrderMeta(reqId)

	case wsRequestCache.CANCEL_ORDER:

		switch usage {
		case usageEnum.TO_UPBIT_CANCEL_TRANSFER:
			{
				symbolIndex := toUpBitListDataAfter.TrigSymbolIndex
				if symbolIndex > 0 {
					toUpbitListChan.SendSpecial(symbolIndex, toUpbitListChan.Special{
						SigType:      toUpbitListChan.CancelOrderReturn,
						AccountKeyId: accountKeyId,
					})
				}
			}
		case usageEnum.TO_UPBIT_MAIN, usageEnum.TO_UPBIT_PRE:
			{
				if reqOk {
					return
				}
				dynamicLog.Error.GetLog().Errorf("取消订单失败:%s", string(data))
			}
		default:
			dynamicLog.Error.GetLog().Errorf("CANCEL_ORDER: unknown usage %v", usage)
		}

	case wsRequestCache.MODIFY_ORDER:
		{
			switch usage {
			case usageEnum.TO_UPBIT_PRE:
				{
					// 预挂单改单成功,预期之内的行为
					if reqOk {
						return
					}
					toUpBitDataStatic.DyLog.GetLog().Error("预挂单改单失败:", string(data))
				}
			default:
				dynamicLog.Error.GetLog().Errorf("MODIFY_ORDER: unknown orderFrom %v", usage)
			}
		}
	case wsRequestCache.QUERY_ACCOUNT_BALANCE:
		{
			if reqOk {
				switch usage {
				case usageEnum.TO_UPBIT_CANCEL_TRANSFER:
					{
						value := gjson.GetBytes(data, `result.#(asset=="USDT").maxWithdrawAmount`)
						if value.Exists() {
							symbolIndex := toUpBitListDataAfter.TrigSymbolIndex
							if symbolIndex > 0 {
								toUpbitListChan.SendSpecial(symbolIndex, toUpbitListChan.Special{
									Amount:       decimal.RequireFromString(value.String()),
									SigType:      toUpbitListChan.QUERY_ACCOUNT_RETURN,
									AccountKeyId: accountKeyId,
								})
							}
						} else {
							dynamicLog.Error.GetLog().Errorf("QUERY_ACCOUNT_BALANCE: json异常 %s", string(data))
						}
					}
				default:
					dynamicLog.Error.GetLog().Errorf("QUERY_ACCOUNT_BALANCE: unknown usage %v", usage)
				}
			}
		}
	default:
		dynamicLog.Error.GetLog().Errorf("WS_REQUEST: unknown WsRequestType %v", wsMeta.ReqType)
	}
}

// {"reqId":"1763011501780691","retCode":10001,"retMsg":"position idx not match position mode"

// {
//   "reqId": "1763005546197794",
//   "retCode": 0,
//   "retMsg": "OK",
//   "op": "order.create",
//   "data": {
//     "orderId": "e21c2cfd-4753-44c9-a865-38ebe7fe91e4",
//     "orderLinkId": "1763005546197794"
//   },
//   "retExtInfo": {},
//   "header": {
//     "X-Bapi-Limit-Status": "19",
//     "X-Bapi-Limit-Reset-Timestamp": "1763005546199",
//     "Traceid": "570b19c082013e0f66b0ae10b104b366",
//     "Timenow": "1763005546199",
//     "X-Bapi-Limit": "20"
//   },
//   "connId": "d3rihevflflovgonnlu0-5nglh"
// }
