package toUpbitBnPayloadParse

import (
	"time"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/usageEnum"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/orderStatic"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitPointPreBn"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListChan"
	"upbitBnServer/pkg/utils/convertx/byteConvert"
)

func OnTradeLite(data []byte, clientOrderId systemx.WsId16B, meta orderStatic.StaticMeta, pStart, pEnd uint16, accountKeyId uint8) {
	usage := meta.UsageFrom
	switch usage {
	case usageEnum.TO_UPBIT_PRE:
		{
			// 卖出开仓成交,主要是用来驱动策略触发
			if meta.OrderMode != execute.SELL_OPEN_LIMIT {
				return
			}
			// 该笔订单已经被处理过了
			if _, ok := toUpbitPointPreBn.ClientOrderNotOpen.Load(clientOrderId); ok {
				return
			}
			toUpbitPointPreBn.ClientOrderNotOpen.Store(clientOrderId, struct{}{})

			toUpbitListChan.SendTradeLite(meta.SymbolIndex, toUpbitListChan.TrigOrderInfo{
				ClientOrderId: clientOrderId,
				T:             byteConvert.BytesToInt64(data[40:53]),
				P:             byteConvert.PriceByteArrToUint64(data[pStart:pEnd], 8),
			})
			toUpBitDataStatic.SigLog.GetLog().Infof("[%d][%s]触发前成交,%s", accountKeyId, string(clientOrderId[:]), string(data))
			go func() {
				time.Sleep(2 * time.Second)
				orderStatic.GetService().DelOrderMeta(clientOrderId) // 删除所有成交的订单
			}()
		}
	case usageEnum.TO_UPBIT_MAIN:
		{
			//暂时不处理
		}
	default:
		dynamicLog.Error.GetLog().Errorf("TRADE_LITE: unknown usage %v", usage)
	}
}
