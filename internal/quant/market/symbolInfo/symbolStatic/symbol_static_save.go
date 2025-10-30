package symbolStatic

import (
	"upbitBnServer/internal/infra/errorx"
	"upbitBnServer/internal/infra/errorx/errCode"
	"upbitBnServer/internal/infra/errorx/errDefine"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/pkg/utils/convertx"
)

type StaticSave struct {
	SymbolKey   string                    `json:"symbol_key"`
	SymbolName  string                    `json:"symbol_name"`
	SymbolKeyId uint64                    `json:"symbol_key_id"`
	TradeId     uint32                    `json:"trade_id"`
	QuoteId     uint16                    `json:"quote_id"`
	ExType      exchangeEnum.ExchangeType `json:"ex_type"`
	AcType      exchangeEnum.AccountType  `json:"ac_type"`
}

func (s StaticSave) TypeName() string {
	return "StaticSave"
}

func (s StaticSave) Check() error {
	if s.SymbolKeyId == 0 {
		return errDefine.ValueInvalid.WithMetadata(map[string]string{"symbol_key_id": convertx.ToString(s.SymbolKeyId)})
	}
	if s.SymbolKey == "" {
		return errDefine.ValueInvalid.WithMetadata(map[string]string{"symbol_key": s.SymbolKey})
	}
	if s.SymbolName == "" {
		return errDefine.ValueInvalid.WithMetadata(map[string]string{"symbol_name": s.SymbolName})
	}
	if s.TradeId == 0 {
		return errDefine.ValueInvalid.WithMetadata(map[string]string{"trade_id": convertx.ToString(s.TradeId)})
	}
	if s.QuoteId == 0 {
		return errDefine.ValueInvalid.WithMetadata(map[string]string{"quote_id": convertx.ToString(s.QuoteId)})
	}
	return nil
}

// GetNotSupportError 工具方法：构造不支持错误信息
func (s StaticSave) GetNotSupportError(flag string) error {
	return errorx.Newf(errCode.STATIC_NOT_SUPPORTED, "[%s_%s] %s ", s.ExType, s.AcType, flag)
}
