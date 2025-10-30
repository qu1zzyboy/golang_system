package symbolInfo

import "upbitBnServer/internal/quant/exchanges/exchangeEnum"

const (
	ReasonSymbolInfoHttpError = "SYMBOL_INFO_HTTP_ERROR" // 获取交易对信息HTTP错误
	MsgSymbolInfoHttpError    = "获取交易对信息HTTP错误"
)

// GetSymbolKey 获取交易对的symbolKey,直接+效率最高,做了内联优化
func GetSymbolKey(exType exchangeEnum.ExchangeType, acType exchangeEnum.AccountType, symbolName string) string {
	return exType.String_() + acType.String_() + symbolName
}

func MakeSymbolKey4(exType exchangeEnum.ExchangeType, acType exchangeEnum.AccountType, cmcId uint32, quoteId uint16) uint64 {
	return (uint64(exType) << 56) | (uint64(acType) << 48) | (uint64(cmcId) << 16) | uint64(quoteId)
}

func MakeSymbolId(cmcId uint32, quoteId uint16) uint64 {
	return (uint64(cmcId) << 16) | uint64(quoteId)
}

func MakeSymbolKey3(exType exchangeEnum.ExchangeType, acType exchangeEnum.AccountType, symbolId uint64) uint64 {
	return (uint64(exType) << 56) | (uint64(acType) << 48) | symbolId
}

func ParseSymbolKey(key uint64) (exType exchangeEnum.ExchangeType, acType exchangeEnum.AccountType, cmcId uint32, quoteId uint16) {
	exType = exchangeEnum.ExchangeType(key >> 56)
	acType = exchangeEnum.AccountType((key >> 48) & 0xff)
	cmcId = uint32((key >> 16) & 0xffffffff)
	quoteId = uint16(key & 0xffff)
	return
}

func ParseSymbolId(symbolId uint64) (cmcId uint32, quoteId uint16) {
	cmcId = uint32(symbolId >> 16)
	quoteId = uint16(symbolId & 0xffff)
	return
}
