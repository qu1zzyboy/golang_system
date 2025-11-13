package toUpbitBnWsParse

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
	hasMonitorSecHigh int64 // 已经监控的价格
	hasMonitorSecLow  int64 // 已经监控的价格
}

func (s *Parser) Parse(data []byte, wsMeta wsRequestCache.WsRequestMeta, reqId systemx.WsId16B, accountKeyId uint8) {
	reqOk := data[34] == '2' && data[35] == '0' && data[36] == '0'

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
				// 预期之内的行为
				return
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
