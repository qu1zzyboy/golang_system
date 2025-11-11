package symbolLimit

import (
	"upbitBnServer/internal/infra/errorx"
	"upbitBnServer/internal/infra/errorx/errCode"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"

	"github.com/shopspring/decimal"
)

var (
	ErrDynamicNotFound = errorx.New(errCode.DYNAMIC_SYMBOL_NOT_FOUND, "限制交易对信息未找到")
)

type LimitSymbol struct {
	SymbolKey        string          `json:"symbol_key"`
	UpLimitPercent   decimal.Decimal `json:"up_limit_percent"`   // 涨停百分比,bn是1.05
	DownLimitPercent decimal.Decimal `json:"down_limit_percent"` // 跌停百分比
}

func (d LimitSymbol) Check() *errorx.Error {
	if d.UpLimitPercent.IsZero() {
		return errorx.Newf(errCode.INVALID_UP_LIMIT_PERCENT, "涨停百分比:%s", d.UpLimitPercent)
	}
	if d.DownLimitPercent.IsZero() {
		return errorx.Newf(errCode.INVALID_DOWN_LIMIT_PERCENT, "跌停百分比:%s", d.DownLimitPercent)
	}
	return nil
}

func (d LimitSymbol) equal(other LimitSymbol) bool {
	return d.UpLimitPercent.Equal(other.UpLimitPercent) && d.DownLimitPercent.Equal(other.DownLimitPercent)
}

func (d LimitSymbol) PrintMe() {
	dynamicLog.Log.GetLog().Infof("限制交易对信息:涨停百分比:%s,跌停百分比:%s", d.UpLimitPercent, d.DownLimitPercent)
}
