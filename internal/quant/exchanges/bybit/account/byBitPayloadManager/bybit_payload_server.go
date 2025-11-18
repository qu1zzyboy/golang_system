package byBitPayloadManager

import (
	"context"
	"upbitBnServer/internal/quant/account/accountConfig"
	"upbitBnServer/internal/quant/exchanges/bybit/account/byBitPayload"
	"upbitBnServer/internal/quant/exchanges/bybit/account/bybitAccountAvailable"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/pkg/utils/byteUtils"

	"github.com/tidwall/gjson"
)

type Payload struct {
	payload      *byBitPayload.ByBitPayload // payload处理器
	accountKeyId uint8                      // 账户序号
}

func newPayload() *Payload {
	return &Payload{}
}

func (s *Payload) init(ctx context.Context, v accountConfig.Config) error {
	s.accountKeyId = v.AccountId
	s.payload = byBitPayload.NewByBitPayload(v.ApiKeyHmac, v.SecretHmac)
	if err := s.payload.RegisterReadHandler(ctx, v.AccountId, s.OnPayload); err != nil {
		return err
	}
	return nil
}

func (s *Payload) OnPayload(data []byte) {
	switch {
	case data[10] == 'o' && data[11] == 'r':
		// order.linear
		s.onPayloadOrder(data)
	case data[10] == 'e' && data[11] == 'x':
		// execution.fast.linear
		s.onTradeLite(data)
	default:
		if data[2] == 'i' && data[3] == 'd' {
			totalLen := uint16(len(data))
			var id_begin uint16 = 7
			id_end := byteUtils.FindNextQuoteIndex(data, id_begin, totalLen)

			topic_start := id_end + 11
			if data[topic_start] == 'w' && data[topic_start+1] == 'a' && data[topic_start+2] == 'l' {
				bybitAccountAvailable.GetManager().SetAvailable(s.accountKeyId, gjson.GetBytes(data, "data.0.totalAvailableBalance").Float())
				return
			}
		}

		if gjson.GetBytes(data, "op").Exists() {
			// {"req_id":"1763012998605","success":true,"ret_msg":"","op":"auth","conn_id":"d2a4c6evqclvsgos5bjg-1zauw0"}
			// {"req_id":"1763012998605","success":true,"ret_msg":"","op":"subscribe","conn_id":"d2a4c6evqclvsgos5bjg-1zauw0"}
			return
		}
		toUpBitDataStatic.DyLog.GetLog().Errorf("[%d]未知事件类型: %s", s.accountKeyId, string(data))
	}
}
