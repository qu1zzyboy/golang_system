package symbolStatic

import (
	"upbitBnServer/internal/infra/errorx"
	"upbitBnServer/internal/infra/errorx/errCode"
	"upbitBnServer/internal/infra/errorx/errDefine"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/quant/market/symbolInfo"
	"upbitBnServer/pkg/utils/convertx"
)

type StaticTrade struct {
	SymbolName  string                    `json:"symbol_name"`
	SymbolKeyId uint64                    `json:"symbol_key_id"`
	TradeId     uint32                    `json:"trade_id"`
	QuoteId     uint16                    `json:"quote_id"`
	ExType      exchangeEnum.ExchangeType `json:"ex_type"`
	AcType      exchangeEnum.AccountType  `json:"ac_type"`
}

// 全量的所有可交易品种静态信息

// GetSymbolKey 工具方法：获取 SymbolKeyId(优先缓存)
func (s StaticTrade) GetSymbolKey() uint64 {
	if s.SymbolKeyId != 0 {
		return s.SymbolKeyId
	}
	return symbolInfo.MakeSymbolKey4(s.ExType, s.AcType, s.TradeId, s.QuoteId)
}

func (s StaticTrade) TypeName() string {
	return "StaticTrade"
}

func (s StaticTrade) Check() error {
	if s.SymbolKeyId == 0 {
		return errDefine.ValueInvalid.WithMetadata(map[string]string{"symbol_key_id": convertx.ToString(s.SymbolKeyId)})
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
func (s StaticTrade) GetNotSupportError(flag string) error {
	return errorx.Newf(errCode.STATIC_NOT_SUPPORTED, "[%s_%s] %s ", s.ExType, s.AcType, flag)
}
