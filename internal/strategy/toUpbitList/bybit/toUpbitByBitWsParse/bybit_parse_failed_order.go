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
	"upbitBnServer/pkg/utils/byteUtils"
	"upbitBnServer/pkg/utils/convertx/byteConvert"

	"github.com/shopspring/decimal"
)

// {"reqId":"1763026852291724","retCode":110007,"retMsg":"ab not enough for new order"
// {"reqId":"1763027838672461","retCode":30228,"retMsg":"No new positions during delisting." 币种正在下架，不能新开仓
// "retCode":110007,"retMsg":"CheckMarginRatio fail! InsufficientAB" 可用余额不足
// {"reqId":"1763086933896632","retCode":20006,"retMsg":"Duplicate reqId"
// {"reqId":"C763362917332254","retCode":10006,"retMsg":"Too many visits. Exceeded the API Rate Limit
// {"reqId":"C763515937969044","retCode":110001,"retMsg":"order not exists or too late to cancel"

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
			err_end := byteUtils.FindNextCommaIndex(data, 38, totalLen)
			toUpbitListChan.SendSpecial(symbolIndex, toUpbitListChan.Special{
				Amount:       decimal.Zero,
				ErrCode:      byteConvert.BytesToInt64(data[38:err_end]),
				SigType:      toUpbitListChan.FailureOrder,
				AccountKeyId: accountKeyId,
			})
			/**********过滤掉常见错误，然后打印************/
			toUpBitDataStatic.DyLog.GetLog().Errorf("账户[%d] TO_UPBIT_MAIN 下单失败,请求:%s,返回:%s", accountKeyId, string(wsMeta.ReqJson), string(data))
			notify.GetNotify().SendImportantErrorMsg(map[string]string{"msg": "下单失败", "clientOrderId": string(clientOrderId[:])})
		}
	default:
		dynamicLog.Error.GetLog().Errorf("PLACE_ORDER: unknown orderFrom %v", oMeta.UsageFrom)
	}
}
