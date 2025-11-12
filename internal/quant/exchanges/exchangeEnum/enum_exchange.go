package exchangeEnum

import (
	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/infra/errorx"
	"upbitBnServer/internal/infra/errorx/errCode"
	"upbitBnServer/internal/infra/errorx/errDefine"
	"upbitBnServer/pkg/utils/convertx"
)

type ExchangeType uint8

const (
	BINANCE ExchangeType = iota
	BYBIT
	UPBIT
	TREE_NEWS
)

func (s ExchangeType) GetNotSupportError(flag string) error {
	return errorx.Newf(errCode.ENUM_NOT_SUPPORTED, "EXTYPE_NOT_SUPPORT[%s] %s ", s.String(), flag)
}

func (s ExchangeType) Verify() error {
	switch s {
	case BINANCE, BYBIT:
		return nil
	default:
		return errDefine.EnumDefineError.WithMetadata(map[string]string{
			defineJson.EnumType: "ExchangeType",
			defineJson.Value:    convertx.ToString(s),
		})
	}
}

func (s ExchangeType) String() string {
	switch s {
	case BINANCE:
		return "BINANCE"
	case BYBIT:
		return "BYBIT"
	case UPBIT:
		return "UPBIT"
	case TREE_NEWS:
		return "TREE_NEWS"
	default:
		return "ERROR"
	}
}

func (s ExchangeType) String_() string {
	switch s {
	case BINANCE:
		return "BINANCE_"
	case BYBIT:
		return "BYBIT_"
	case UPBIT:
		return "UPBIT_"
	default:
		return "ERROR_"
	}
}
