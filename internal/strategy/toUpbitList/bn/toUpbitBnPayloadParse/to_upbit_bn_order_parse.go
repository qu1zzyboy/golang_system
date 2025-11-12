package toUpbitBnPayloadParse

import (
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/usageEnum"
	"upbitBnServer/internal/quant/account/bnPayload"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/orderStatic"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitPointPreBn"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListChan"
	"upbitBnServer/pkg/utils/byteUtils"
	"upbitBnServer/pkg/utils/convertx/byteConvert"
)

func OnPayloadOrder(data []byte, clientOrderId systemx.WsId16B, meta orderStatic.StaticMeta, totalLen, cidEnd uint16, accountKeyId uint8) {
	usage := meta.UsageFrom
	switch usage {
	case usageEnum.TO_UPBIT_PRE:
		{
			// 卖出开仓成交,主要是用来驱动策略触发
			if meta.OrderMode != execute.SELL_OPEN_LIMIT {
				return
			}
			q_start := cidEnd + 40
			q_end := byteUtils.FindNextQuoteIndex(data, q_start, totalLen)
			p_start := q_end + 7
			p_end := byteUtils.FindNextQuoteIndex(data, p_start, totalLen)

			//ap的第一个字符为0",说明没有成交
			if data[p_end+8] == '0' && data[p_end+9] == '"' {
				ap_start := p_end + 8
				ap_end := byteUtils.FindNextQuoteIndex(data, ap_start, totalLen)

				x_start := ap_end + 16
				x_end := byteUtils.FindNextQuoteIndex(data, x_start, totalLen)

				switch data[x_end+7] {
				case 'N':
					toUpbitPointPreBn.OnOrderUpdate(true, clientOrderId)
				case 'C':
					toUpbitPointPreBn.OnOrderUpdate(false, clientOrderId)
				case 'R':
					toUpbitPointPreBn.OnOrderUpdate(false, clientOrderId)
				case 'E':
					toUpbitPointPreBn.OnOrderUpdate(false, clientOrderId)
				default:
					toUpBitDataStatic.DyLog.GetLog().Errorf("[%d]ORDER_UPDATE: unknown order status[%d], json: %s", accountKeyId, x_end+7, string(data))
				}
				return
			}
			// 该笔订单已经被处理过了
			if _, ok := toUpbitPointPreBn.ClientOrderNotOpen.Load(clientOrderId); ok {
				return
			}
			toUpbitPointPreBn.ClientOrderNotOpen.Store(clientOrderId, struct{}{})

			toUpbitListChan.SendTradeLite(meta.SymbolIndex, toUpbitListChan.TrigOrderInfo{
				ClientOrderId: clientOrderId,
				T:             byteConvert.BytesToInt64(data[30:43]),
				P:             byteConvert.PriceByteArrToUint64(data[p_start:p_end], 8),
			})
			toUpBitDataStatic.SigLog.GetLog().Infof("[%d][%s]触发前成交,%s", accountKeyId, string(clientOrderId[:]), string(data))
		}
	case usageEnum.TO_UPBIT_MAIN:
		{
			// 1、只管触发标的的订单
			if meta.SymbolIndex != toUpBitListDataAfter.TrigSymbolIndex {
				toUpBitDataStatic.DyLog.GetLog().Errorf("触发后异常订单返回:%s", string(data))
				return
			}
			orderStatus, X_len, ap_start, x_end := bnPayload.ParseOrderStatus(data, cidEnd, totalLen, accountKeyId)
			isOnline := execute.IsOrderOnLine(orderStatus)
			evt := toUpBitListDataAfter.OnSuccessEvt{
				ClientOrderId: clientOrderId,
				IsOnline:      isOnline,
				OrderMode:     meta.OrderMode,
				AccountKeyId:  accountKeyId,
			}
			if !isOnline {
				if !(data[ap_start] == '0' && data[ap_start+1] == '"') {
					//ap的第一个字符不为0",说明有成交
					X_Start := x_end + 7
					X_end := X_Start + X_len

					i_start := X_end + 6
					i_end := byteUtils.FindNextCommaIndex(data, i_start, totalLen)

					l_start := i_end + 6
					l_end := byteUtils.FindNextQuoteIndex(data, l_start, totalLen)

					z_start := l_end + 7
					z_end := byteUtils.FindNextQuoteIndex(data, z_start, totalLen)
					evt.Volume = byteConvert.ByteArrToF64(data[z_start:z_end])
				}
			}
			evt.TimeStamp = byteConvert.BytesToInt64(data[30:43])
			toUpbitListChan.SendSuOrder(meta.SymbolIndex, evt)
		}
	default:
		dynamicLog.Error.GetLog().Errorf("ORDER_UPDATE: unknown orderFrom %v,%s", meta.UsageFrom, string(data))
	}
}
