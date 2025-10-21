package bnPayloadManager

import (
	"context"

	"github.com/hhh500/upbitBnServer/internal/quant/account/accountConfig"
	"github.com/hhh500/upbitBnServer/internal/quant/account/bnPayload"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/bn/toUpBitListBnExecute"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

type Payload struct {
	payload      *bnPayload.BnPayload // payload处理器
	accountKeyId uint8                // 账户序号
}

func newPayload() *Payload {
	return &Payload{}
}

func (s *Payload) init(ctx context.Context, v accountConfig.Config) error {
	s.accountKeyId = v.AccountId
	s.payload = bnPayload.NewBnPayload(v.ApiKeyHmac, v.SecretHmac)
	if err := s.payload.RegisterReadHandler(ctx, v.AccountId, s.OnPayload); err != nil {
		return err
	}
	return nil
}

func (s *Payload) OnPayload(data []byte) {
	eveType := gjson.GetBytes(data, "e").String()
	switch eveType {
	case bnPayload.ORDER_TRADE_UPDATE:
		s.onPayloadOrder(data)
	case bnPayload.TRADE_LITE:
		s.onTradeLite(data)
	case bnPayload.ACCOUNT_UPDATE:
		if s.accountKeyId == 11 {
			wb := gjson.GetBytes(data, `a.B.#(a=="USDT").wb`)
			if wb.Exists() {
				max := decimal.RequireFromString(wb.String())
				if max.GreaterThan(decimal.Zero) {
					toUpBitListBnExecute.GetExecute().OnTransOut(max)
				}
			}
		}
	default:
		if eveType == bnPayload.ALGO_UPDATE || eveType == bnPayload.ACCOUNT_CONFIG_UPDATE {
			return
		}
		toUpBitListDataStatic.DyLog.GetLog().Errorf("[%d]未知事件类型: %s", s.accountKeyId, string(data))
	}
}

// {
// 	"a": "USDT",
// 	"wb": "12.31980017", //钱包余额
// 	"cw": "12.31980017", // 除去逐仓仓位保证金的钱包余额
// 	"bc": "0"
// }
