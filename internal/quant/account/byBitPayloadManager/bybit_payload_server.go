package byBitPayloadManager

import (
	"context"
	"upbitBnServer/internal/quant/account/accountConfig"
	"upbitBnServer/internal/quant/account/byBitPayload"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
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
		toUpBitDataStatic.DyLog.GetLog().Errorf("[%d]未知事件类型: %s", s.accountKeyId, string(data))
	}
}
