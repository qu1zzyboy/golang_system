package toUpbitByBitPayloadParse

import (
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/usageEnum"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/orderStatic"
	"upbitBnServer/internal/strategy/toUpbitList/bybit/toUpbitPointPreByBit"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListChan"
	"upbitBnServer/pkg/utils/byteUtils"
	"upbitBnServer/pkg/utils/convertx/byteConvert"
	"upbitBnServer/pkg/utils/pow10Utils"
)

func OnPayloadOrder(data []byte, clientOrderId systemx.WsId16B, meta orderStatic.StaticMeta, totalLen, o_id_start, o_id_end, cidEnd uint16, accountKeyId uint8) {
	usage := meta.UsageFrom
	switch usage {
	case usageEnum.TO_UPBIT_PRE:
		{
			// 卖出开仓成交,主要是用来驱动策略触发
			if meta.OrderMode != execute.SELL_OPEN_LIMIT {
				return
			}
			var o_start uint16
			switch data[cidEnd+28] {
			case 'B':
				o_start = cidEnd + 64
			case 'S':
				o_start = cidEnd + 65
			default:
				toUpBitDataStatic.DyLog.GetLog().Errorf("B_S 解析异常:%s", string(data))
				return
			}
			switch data[o_start] {
			case 'N':
				toUpbitPointPreByBit.OnOrderUpdate(true, clientOrderId)
				orderStatic.GetService().SaveOrderIdToClientOrderId(string(data[o_id_start:o_id_end]), clientOrderId)
				return
			case 'P':
			case 'F':
			case 'C':
				toUpbitPointPreByBit.OnOrderUpdate(false, clientOrderId)
				return
			case 'R':
				toUpbitPointPreByBit.OnOrderUpdate(false, clientOrderId)
				return
			default:
				toUpBitDataStatic.DyLog.GetLog().Errorf("ORDER_UPDATE: unknown order status[%d], json: %s", o_start, string(data))
			}
			// 该笔订单已经被处理过了
			if _, ok := toUpbitPointPreByBit.ClientOrderNotOpen.Load(clientOrderId); ok {
				return
			}
			toUpbitPointPreByBit.ClientOrderNotOpen.Store(clientOrderId, struct{}{})

			toUpbitListChan.SendTradeLite(meta.SymbolIndex, toUpbitListChan.TrigOrderInfo{
				ClientOrderId: clientOrderId,
				//T:             byteConvert.BytesToInt64(data[30:43]),
				//P:             byteConvert.PriceByteArrToUint64(data[p_start:p_end], 8),
			})
			toUpBitDataStatic.SigLog.GetLog().Infof("[%d][%s]触发前成交,%s", accountKeyId, string(clientOrderId[:]), string(data))
		}
	case usageEnum.TO_UPBIT_MAIN:
		// 1、只管触发标的的订单
		if meta.SymbolIndex != toUpBitListDataAfter.TrigSymbolIndex {
			toUpBitDataStatic.DyLog.GetLog().Errorf("触发后异常订单返回:%s", string(data))
			return
		}
		var o_start uint16
		switch data[cidEnd+28] {
		case 'B':
			o_start = cidEnd + 64
		case 'S':
			o_start = cidEnd + 65
		default:
			toUpBitDataStatic.DyLog.GetLog().Errorf("B_S 解析异常:%s", string(data))
			return
		}
		var orderStatus execute.OrderStatus
		var X_len uint16
		switch data[o_start] {
		case 'N':
			orderStatus = execute.NEW
			X_len = 3
		case 'P':
			orderStatus = execute.PARTIALLY_FILLED
			X_len = 15
		case 'F':
			orderStatus = execute.FILLED
			X_len = 6
		case 'C':
			orderStatus = execute.CANCELED
			X_len = 9
		case 'R':
			orderStatus = execute.REJECTED
			X_len = 8
		default:
			toUpBitDataStatic.DyLog.GetLog().Errorf("[%d]ORDER_UPDATE: unknown order status[%d], json: %s", accountKeyId, o_start, string(data))
			return
		}
		isOnline := execute.IsOrderOnLine(orderStatus)
		evt := toUpBitListDataAfter.OnSuccessEvt{
			ClientOrderId: clientOrderId,
			IsOnline:      isOnline,
			OrderMode:     meta.OrderMode,
			AccountKeyId:  accountKeyId,
		}
		if !isOnline {
			o_end := o_start + X_len
			p_start := o_end + 98
			p_end := byteUtils.FindNextQuoteIndex(data, p_start, totalLen)

			// ","qty":" 9
			q_start := p_end + 9
			q_end := byteUtils.FindNextQuoteIndex(data, q_start, totalLen)

			// ","avgPrice":" 14
			avg_p_start := q_end + 14
			avg_p_end := byteUtils.FindNextQuoteIndex(data, avg_p_start, totalLen)

			// 有成交
			if avg_p_end-avg_p_start > 1 {
				left_q_start := avg_p_end + 15
				left_q_end := byteUtils.FindNextQuoteIndex(data, left_q_start, totalLen)
				left_q := byteConvert.ByteArrToF64(data[left_q_start:left_q_end])
				origVol := float64(meta.Qvalue) / pow10Utils.ToPowF64(uint8(meta.Qscale))
				evt.Volume = origVol - left_q
			}
		} else {
			// evt.TimeStamp = data.Get("createdTime").Int()
		}
		toUpbitListChan.SendSuOrder(meta.SymbolIndex, evt)

	default:
		dynamicLog.Error.GetLog().Errorf("ORDER_UPDATE: unknown orderFrom %v,%s", meta.UsageFrom, string(data))
	}
}
