package bnPayloadManager

import (
	"context"

	"upbitBnServer/internal/quant/account/accountConfig"
	"upbitBnServer/internal/quant/account/bnPayload"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitListBnSymbolArr"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"

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
			// 必须是转入导致的资金变化
			if gjson.GetBytes(data, "a.m").String() != "ADMIN_DEPOSIT" {
				return
			}
			wb := gjson.GetBytes(data, `a.B.#(a=="USDT").wb`)
			if wb.Exists() {
				symbolIndex := toUpBitListDataAfter.TrigSymbolIndex
				if symbolIndex > (-1) {
					max_ := decimal.RequireFromString(wb.String())
					if max_.GreaterThan(decimal.Zero) {
						toUpbitListBnSymbolArr.GetSymbolObj(symbolIndex).OnTransOut(max_)
					}
				}
			}
		}
	default:
		if eveType == bnPayload.ALGO_UPDATE || eveType == bnPayload.ACCOUNT_CONFIG_UPDATE {
			return
		}
		toUpBitDataStatic.DyLog.GetLog().Errorf("[%d]未知事件类型: %s", s.accountKeyId, string(data))
	}
}

// {"e":"ACCOUNT_UPDATE","T":1761103247006,"E":1761103247006,"a":{"B":[{"a":"USDT","wb":"1","cw":"1","bc":"1"}],"P":[],"m":"ADMIN_DEPOSIT"}}
// {"e":"ACCOUNT_UPDATE","T":1761103247282,"E":1761103247282,"a":{"B":[{"a":"USDT","wb":"0","cw":"0","bc":"-1"}],"P":[],"m":"ADMIN_WITHDRAW"}}

// {
// 	"a": "USDT",
// 	"wb": "12.31980017", //钱包余额
// 	"cw": "12.31980017", //除去逐仓仓位保证金的钱包余额
// 	"bc": "0"
// }
