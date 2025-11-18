package toUpbitBnWsParse

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
	"upbitBnServer/pkg/utils/byteUtils"
	"upbitBnServer/pkg/utils/convertx/byteConvert"

	"github.com/shopspring/decimal"
)

func (s *Parser) onPlaceOrderFailed(data []byte, wsMeta wsRequestCache.WsRequestMeta, clientOrderId systemx.WsId16B, accountKeyId uint8) {
	totalLen := uint16(len(data))
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
			switch {
			case data[55] == '2' && data[56] == '0' && data[57] == '1' && data[58] == '3':
				// "code":-2013,"msg":"Order does not exist."
				return
			}
			// "code":-4116,"msg":"ClientOrderId is duplicated."
			dynamicLog.Error.GetLog().Errorf("[%d]下单失败: 请求:%s,返回%s", accountKeyId, string(wsMeta.ReqJson), string(data))
		}
	case usageEnum.TO_UPBIT_MONITOR:
		{
			if data[55] == '4' && data[56] == '0' && data[57] == '1' && data[58] == '6' {
				s.process4016(clientOrderId, data, symbolIndex, totalLen)
				return
			}
		}
	case usageEnum.TO_UPBIT_MAIN:
		{
			// 需要判断是不是触发的币种
			if symbolIndex != toUpBitListDataAfter.TrigSymbolIndex {
				/*
					一般都是因为网络太慢了,推送过来太慢了
				*/
				if toUpBitListDataAfter.TrigSymbolIndex == (-1) {
					switch {
					case data[55] == '2' && data[56] == '0' && data[57] == '1' && data[58] == '9':
						// "code":-2019,"msg":"Margin is insufficient."
						return
					case data[55] == '2' && data[56] == '0' && data[57] == '2' && data[58] == '7':
						// "code":-2027,"msg":"Exceeded the maximum allowable position at current leverage."
						return
					case data[55] == '5' && data[56] == '0' && data[57] == '2' && data[58] == '2':
						// "code":-5022,"msg":"Due to the order could not be executed as maker
						return
					case data[55] == '4' && data[56] == '0' && data[57] == '1' && data[58] == '6':
						// "code":-4016,"msg":"Limit price can't be higher than 0.0012859."
						return
					}
				}
				toUpBitDataStatic.DyLog.GetLog().Errorf("[%d]触发后异常订单,请求数据:%s,返回数据%s", accountKeyId, string(wsMeta.ReqJson), string(data))
				return
			}

			// 接下来是正在触发的订单的错误返回
			if data[55] == '4' && data[56] == '0' && data[57] == '1' && data[58] == '6' {
				s.process4016(clientOrderId, data, symbolIndex, totalLen)
				return
			}
			// 发送错误码
			toUpbitListChan.SendSpecial(symbolIndex, toUpbitListChan.Special{
				Amount:       decimal.Zero,
				ErrCode:      byteConvert.BytesToInt64([]byte{data[55], data[56], data[57], data[58]}),
				SigType:      toUpbitListChan.FailureOrder,
				AccountKeyId: accountKeyId,
			})
			/**********过滤掉常见错误，然后打印************/
			switch {
			case data[55] == '2' && data[56] == '0' && data[57] == '1' && data[58] == '9':
				// "code":-2019,"msg":"Margin is insufficient."
				return
			case data[55] == '5' && data[56] == '0' && data[57] == '2' && data[58] == '2':
				// "code":-5022,"msg":"Due to the order could not be executed as maker
				return
			case data[55] == '2' && data[56] == '0' && data[57] == '2' && data[58] == '7':
				// "code":-2027,"msg":"Exceeded the maximum allowable position at current leverage."
				return
			}
			toUpBitDataStatic.DyLog.GetLog().Errorf("账户[%d]下单失败,请求:%s,返回:%s", accountKeyId, string(wsMeta.ReqJson), string(data))
			notify.GetNotify().SendImportantErrorMsg(map[string]string{"msg": "下单失败", "clientOrderId": string(clientOrderId[:])})
		}
	default:
		dynamicLog.Error.GetLog().Errorf("PLACE_ORDER: unknown orderFrom %v", oMeta.UsageFrom)
	}
}

func (s *Parser) process4016(clientOrderId [16]byte, data []byte, symbolIndex systemx.SymbolIndex16I, totalLen uint16) {

	// {"id":"P761995167005637","status":400,"error":{"code":-4024,"msg":"Limit price can't be lower than 3684.93."}}
	monitorSec := byteConvert.BytesToInt64(clientOrderId[:10])
	if monitorSec <= s.hasMonitorSecHigh {
		return
	}
	s.hasMonitorSecHigh = monitorSec
	lastPointIndex := totalLen - 4
	// 出现了异常数据,重新获取最后一个点出现的索引
	if data[lastPointIndex] != '.' {
		lastPointIndex = byteUtils.FindLastPointIndex(data, totalLen-1)
	}
	priceBeginIndex := byteUtils.FindLastSpaceIndex(data, lastPointIndex)
	priceBuxMax := byteConvert.ByteArrToF64(data[priceBeginIndex:lastPointIndex])
	toUpbitListChan.SendMonitorData(symbolIndex, toUpbitListChan.MonitorResp{P: priceBuxMax})
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
