package toUpbitBybitSymbol

import (
	"time"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/pkg/utils/byteUtils"
	"upbitBnServer/pkg/utils/convertx/byteConvert"
)

func (s *Single) onAggTrade(b []byte) {
	/****处理成交数据****/
	if toUpBitListDataAfter.LoadTrig() {
		/*********************上币已经触发**************************/
		if s.symbolIndex != toUpBitListDataAfter.TrigSymbolIndex {
			return
		}
	} else {
		/*********************上币还未触发**************************/
		e_begin := 54 - 7 + s.symbolLen
		e_end := e_begin + 13
		msE := byteConvert.BytesToInt64(b[e_begin:e_end])
		t_begin := e_end + 14
		t_end := t_begin + 13
		msT := byteConvert.BytesToInt64(b[t_begin:t_end])
		// 数据太旧则丢弃
		if msT <= s.committedTs {
			s.agLatencyTotal.Record(s.symbolName, float64(time.Now().UnixMicro()-1000*msE)) // 记录总延迟
			return
		}
		byteLen := uint16(len(b))
		s_end := t_end + s.symbolLen + 6
		S_begin := s_end + 7
		S_end := S_begin + 4
		if b[S_begin] == 'B' {
			S_end = S_begin + 3
		}
		q_begin := S_end + 7
		q_end := byteUtils.FindNextQuoteIndex(b, q_begin, byteLen)

		p_begin := q_end + 7
		p_end := byteUtils.FindNextQuoteIndex(b, p_begin, byteLen)

		priceU8 := byteConvert.PriceByteArrToUint64(b[p_begin:p_end], 8)
		// 2、计算触发信号
		s.checkMarket(msT, priceU8)
		s.agLatencyTotal.Record(s.symbolName, float64(time.Now().UnixMicro()-1000*msE))
	}
}

//{"topic":"publicTrade.ETHUSDT","type":"snapshot","ts":  54

// {
//   "topic": "publicTrade.ETHUSDT",
//   "type": "snapshot",
//   "ts": 1762310912746,
//   "data": [
//     {
//       "T": 1762310912745,
//       "s": "ETHUSDT",
//       "S": "Buy", //吃單方向
//       "v": "0.35", //成交數量
//       "p": "3299.34",
//       "L": "ZeroMinusTick", //價格變化的方向
//       "i": "44b10f83-153f-5222-a6d3-64875e37df15", //成交Id
//       "BT": false, //成交類型是否為大宗交易
//       "RPI": false, //成交類型是否為RPI交易
//       "seq": 326619617914  //撮合序列號
//     },
//     {
//       "T": 1762310912745,
//       "s": "ETHUSDT",
//       "S": "Buy",
//       "v": "0.12",
//       "p": "3299.34",
//       "L": "ZeroMinusTick",
//       "i": "b45adabe-9e54-5779-bd18-60cf7a182f10",
//       "BT": false,
//       "RPI": false,
//       "seq": 326619617914
//     }
//   ]
// }
