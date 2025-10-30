package exchangeEnum

import (
	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/infra/errorx"
	"upbitBnServer/internal/infra/errorx/errCode"
	"upbitBnServer/internal/infra/errorx/errDefine"
	"upbitBnServer/pkg/utils/convertx"
)

type AccountType uint8

const (
	SPOT AccountType = iota
	FUTURE
	SWAP
	FULL_MARGIN
	ISOLATED_MARGIN
)

func (s AccountType) GetNotSupportError(flag string) error {
	return errorx.Newf(errCode.ENUM_NOT_SUPPORTED, "ACTYPE_NOT_SUPPORT[%s] %s ", s.String(), flag)
}

func (s AccountType) Verify() error {
	switch s {
	case SPOT, FUTURE, SWAP:
		return nil
	default:
		return errDefine.EnumDefineError.WithMetadata(map[string]string{
			defineJson.EnumType: "AccountType",
			defineJson.Value:    convertx.ToString(s),
		})
	}
}

func (s AccountType) String() string {
	switch s {
	case SPOT:
		return "SPOT"
	case FUTURE:
		return "FUTURE"
	default:
		return "ERROR"
	}
}

func (s AccountType) String_() string {
	switch s {
	case SPOT:
		return "SPOT_"
	case FUTURE:
		return "FUTURE_"
	default:
		return "ERROR_"
	}
}
