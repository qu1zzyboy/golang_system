package bnDriverWsParse

import (
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/quant/execute/order/orderStaticMeta"
	"upbitBnServer/internal/quant/execute/order/wsRequestCache"
	"upbitBnServer/internal/strategy/newsDrive/common/driverListChan"
	"upbitBnServer/internal/strategy/newsDrive/common/driverOrderDedup"
	"upbitBnServer/internal/strategy/newsDrive/common/driverStatic"
	"upbitBnServer/pkg/singleton"
	"upbitBnServer/server/usageEnum"

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
}

func (s *Parser) Parse(data []byte, wsMeta *wsRequestCache.WsRequestMeta, reqId systemx.WsId16B, accountKeyId uint8, reqOk bool) {
	usage := wsMeta.UsageFrom

	switch wsMeta.ReqType {
	case wsRequestCache.PLACE_ORDER:

		/* 接收到返回就算这个请求是有效的,不管是成功还是失败 */
		driverOrderDedup.PlaceManager.Delete(wsMeta.ClientOrderId)

		oMeta, ok := orderStaticMeta.GetService().GetOrderMeta(wsMeta.ClientOrderId)
		// 不属于这个服务的订单直接pass
		if !ok {
			dynamicLog.Error.GetLog().Errorf("[%d]PLACE_ORDER: [%s] orderFrom not found %s", accountKeyId, string(wsMeta.ClientOrderId[:]), string(data))
			return
		}

		//下单成功由payload处理,这里只处理下单失败
		if reqOk {
			evt := orderModel.OnSuccessEvt{
				ClientOrderId: wsMeta.ClientOrderId,
				T:             gjson.GetBytes(data, "T").Int(),
				OrderMode:     oMeta.OrderMode,
				AccountKeyId:  accountKeyId,
				UsageFrom:     usage,
				OrderStatus:   execute.NEW,
			}
			driverListChan.SendSuOrder(oMeta.SymbolIndex, evt)
			return
		}
		s.onPlaceOrderFailed(data, wsMeta, accountKeyId)
		// 如果是下单失败, 删除 orderStaticMeta
		orderStaticMeta.GetService().DelOrderMeta(reqId)

	case wsRequestCache.MODIFY_ORDER:
		{
			/* 接收到返回就算这个请求是有效的,不管是成功还是失败 */
			driverOrderDedup.ModifyManager.Delete(wsMeta.ClientOrderId)

			switch usage {
			case usageEnum.NEWS_DRIVE_PRE:
				{
					// 预挂单改单成功,预期之内的行为
					if reqOk {
						return
					}
					driverStatic.DyLog.GetLog().Errorf("预挂单改单失败,请求:%s,响应:%s", string(wsMeta.ReqJson), string(data))
				}
			default:
				dynamicLog.Error.GetLog().Errorf("MODIFY_ORDER: unknown orderFrom %v", usage)
			}
		}

	case wsRequestCache.CANCEL_ORDER:

		/* 接收到返回就算这个请求是有效的,不管是成功还是失败 */
		driverOrderDedup.CancelManager.Delete(wsMeta.ClientOrderId)

		switch usage {
		case usageEnum.CANCEL_AND_TRANSFER:
			{
				symbolIndex := driverStatic.TrigSymbolIndex
				if symbolIndex > 0 {
					driverListChan.SendSpecial(symbolIndex, driverListChan.Special{
						SigType:      driverListChan.CANCEL_ORDER_RETURN,
						AccountKeyId: accountKeyId,
					})
				}
			}
		case usageEnum.NEWS_DRIVE_MAIN, usageEnum.NEWS_DRIVE_PRE:
			{
				// 预期之内的行为
				return
			}
		default:
			dynamicLog.Error.GetLog().Errorf("CANCEL_ORDER: unknown usage %v", usage)
		}

	case wsRequestCache.QUERY_ACCOUNT_BALANCE:
		{
			if reqOk {
				switch usage {
				case usageEnum.CANCEL_AND_TRANSFER:
					{
						value := gjson.GetBytes(data, `result.#(asset=="USDT").maxWithdrawAmount`)
						if value.Exists() {
							symbolIndex := driverStatic.TrigSymbolIndex
							if symbolIndex > 0 {
								driverListChan.SendSpecial(symbolIndex, driverListChan.Special{
									Amount:       decimal.RequireFromString(value.String()),
									SigType:      driverListChan.QUERY_ACCOUNT_RETURN,
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
