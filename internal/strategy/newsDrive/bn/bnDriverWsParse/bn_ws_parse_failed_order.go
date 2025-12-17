package bnDriverWsParse

import (
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/observe/notify"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/quant/execute/order/orderStaticMeta"
	"upbitBnServer/internal/quant/execute/order/wsRequestCache"
	"upbitBnServer/internal/strategy/newsDrive/common/driverListChan"
	"upbitBnServer/internal/strategy/newsDrive/common/driverStatic"
)

func (s *Parser) onPlaceOrderFailed(data []byte, wsMeta *wsRequestCache.WsRequestMeta, accountKeyId uint8) {
	totalLen := uint16(len(data))
	oMeta, ok := orderStaticMeta.GetService().GetOrderMeta(wsMeta.ClientOrderId)
	// 不属于这个服务的订单直接pass
	if !ok {
		dynamicLog.Error.GetLog().Errorf("[%d]PLACE_ORDER: [%s] orderFrom not found %s", accountKeyId, wsMeta.ClientOrderId, string(data))
		return
	}
	symbolIndex := oMeta.SymbolIndex
	usage := wsMeta.UsageFrom

	// 所有的下单失败都应该被消费到
	var fa orderModel.OnFailedEvt
	fa.ClientOrderId = wsMeta.ClientOrderId
	fa.ErrorCode = int32(data[55]-'0')*1000 + int32(data[56]-'0')*100 + int32(data[57]-'0')*10 + int32(data[58]-'0')
	fa.UsageFrom = usage
	fa.ReqType = wsMeta.ReqType
	fa.AccountKeyId = accountKeyId

	// 4016是特殊错误，优先返回
	if fa.ErrorCode == 4016 {
		fa.P, ok = s.process4016(byteConvert.BytesToInt64(wsMeta.ClientOrderId[:10]), data, totalLen)
		// 更新成功
		if ok {
			driverListChan.SendFaOrder(symbolIndex, fa)
			driverStatic.DyLog.GetLog().Info(string(data))
			return
		}
	}
	if fa.ErrorCode == 4024 {
		fa.P, ok = s.process4024(byteConvert.BytesToInt64(wsMeta.ClientOrderId[:10]), data, totalLen)
		// 更新成功
		if ok {
			driverListChan.SendFaOrder(symbolIndex, fa)
			driverStatic.DyLog.GetLog().Info(string(data))
			return
		}
	}

	switch usage {
	case usageEnum.NEWS_DRIVE_PRE:

		switch fa.ErrorCode {
		case 2013: // "code":-2013,"msg":"Order does not exist."
		default:
			dynamicLog.Error.GetLog().Errorf("[%d]下单失败: 请求:%s,返回%s", accountKeyId, string(wsMeta.ReqJson), string(data))
		}
		// "code":-4116,"msg":"ClientOrderId is duplicated."

	case usageEnum.NEWS_DRIVE_MAIN:
		{
			/**********过滤掉常见错误，然后打印************/
			switch fa.ErrorCode {
			case 4016:
			case 2019: //"code":-2019,"msg":"Margin is insufficient."
			case 5022: //"code":-5022,"msg":"Due to the order could not be executed as maker
			case 2027: //"code":-2027,"msg":"Exceeded the maximum allowable position at current leverage.
			default:
				driverStatic.DyLog.GetLog().Errorf("账户[%d]下单失败,请求:%s,返回:%s", accountKeyId, string(wsMeta.ReqJson), string(data))
				notify.GetNotify().SendImportantErrorMsg(map[string]string{"msg": string(data), "clientOrderId": string(wsMeta.ClientOrderId[:])})
			}
		}
	// 5%触发后探针是否挂单成功
	case usageEnum.POINTER_ASKS_5,
		usageEnum.POINTER_ASKS_6,
		usageEnum.POINTER_ASKS_7,
		usageEnum.POINTER_ASKS_8,
		usageEnum.POINTER_ASKS_9,
		usageEnum.POINTER_ASKS_10,
		usageEnum.POINTER_ASKS_11,
		usageEnum.POINTER_ASKS_12,
		usageEnum.POINTER_ASKS_13,
		usageEnum.POINTER_ASKS_14,
		usageEnum.POINTER_ASKS_15,
		usageEnum.POINTER_BIDS_1,
		usageEnum.POINTER_BIDS_2,
		usageEnum.POINTER_BIDS_3,
		usageEnum.POINTER_BIDS_4,
		usageEnum.POINTER_BIDS_5,
		usageEnum.POINTER_BIDS_6,
		usageEnum.POINTER_BIDS_7,
		usageEnum.POINTER_BIDS_8,
		usageEnum.POINTER_BIDS_9,
		usageEnum.POINTER_BIDS_10,
		usageEnum.POINTER_BIDS_11,
		usageEnum.POINTER_BIDS_12,
		usageEnum.POINTER_BIDS_13,
		usageEnum.POINTER_BIDS_14,
		usageEnum.POINTER_BIDS_15:

		// 接下来是正在触发的探针的错误返回
		switch fa.ErrorCode {
		case 4016:
		case 4024:
		case 5022: //"code":-5022,"msg":"Due to the order could not be executed as maker, the Post Only order will be rejected. The order will not be recorded in the order history"}}
			fa.PointType = pointRespEnum.LIMIT_MAKER_FAILED
		case 4400: //"code":-4400,"msg":"Futures Trading Quantitative Rules violated, only reduceOnly order is allowed, please try again later."}}
			fa.PointType = pointRespEnum.QUANT_VIOLATED
		case 2019: //"code":-2019,"msg":"Margin is insufficient."
		default:
			driverStatic.DyLog.GetLog().Errorf("账户[%d] trig触发后探针%s 下单失败,请求:%s,返回:%s", accountKeyId, usage.String(), string(wsMeta.ReqJson), string(data))
			notify.GetNotify().SendImportantErrorMsg(map[string]string{"msg": "探针失败", "clientOrderId": string(wsMeta.ClientOrderId[:])})
		}
	default:
		dynamicLog.Error.GetLog().Errorf("PLACE_ORDER: unknown orderFrom %v", oMeta.UsageFrom)
	}
	driverListChan.SendFaOrder(symbolIndex, fa)
}

// "code":-4002,"msg":"Price greater than max price."

func (s *Parser) process4016(monitorSec int64, data []byte, totalLen uint16) (priceBuxMax float64, ok bool) {
	// {"id":"P761995167005637","status":400,"error":{"code":-4024,"msg":"Limit price can't be lower than 3684.93."}}
	if monitorSec <= s.hasMonitorSecHigh {
		return 0, false
	}
	s.hasMonitorSecHigh = monitorSec
	lastPointIndex := totalLen - 4
	// 出现了异常数据,重新获取最后一个点出现的索引
	if data[lastPointIndex] != '.' {
		lastPointIndex = byteUtils.FindLastPointIndex(data, totalLen-1)
	}
	priceBeginIndex := byteUtils.FindLastSpaceIndex(data, lastPointIndex)
	priceBuxMax = byteConvert.ByteArrToF64(data[priceBeginIndex:lastPointIndex])
	return priceBuxMax, true
}
