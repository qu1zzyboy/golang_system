package symbolDynamic

import (
	"upbitBnServer/internal/infra/errorx"
	"upbitBnServer/internal/infra/errorx/errCode"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"

	"github.com/shopspring/decimal"
)

var (
	ErrDynamicNotFound = errorx.New(errCode.DYNAMIC_SYMBOL_NOT_FOUND, "动态交易对信息未找到")
)

type DynamicSymbol struct {
	SymbolKey        string          `json:"symbol_key"`
	UpLimitPercent   decimal.Decimal `json:"up_limit_percent"`   // 涨停百分比,bn是1.05
	DownLimitPercent decimal.Decimal `json:"down_limit_percent"` // 跌停百分比
	MinQty           decimal.Decimal `json:"min_qty"`            // 最小下单金额
	TickSize         decimal.Decimal `json:"tick_size"`          // 最小价格变动单位
	LotSize          decimal.Decimal `json:"lot_size"`           // 最小交易单位
	PScale           int32           `json:"p_scale"`
	QScale           int32           `json:"q_scale"`
}

func (d DynamicSymbol) Check() *errorx.Error {
	if d.UpLimitPercent.IsZero() {
		return errorx.Newf(errCode.INVALID_UP_LIMIT_PERCENT, "涨停百分比:%s", d.UpLimitPercent)
	}
	if d.DownLimitPercent.IsZero() {
		return errorx.Newf(errCode.INVALID_DOWN_LIMIT_PERCENT, "跌停百分比:%s", d.DownLimitPercent)
	}
	if d.MinQty.IsZero() {
		return errorx.Newf(errCode.INVALID_MIN_QTY, "最小下单金额:%s", d.MinQty)
	}
	if d.LotSize.IsZero() {
		return errorx.Newf(errCode.INVALID_LOT_SIZE, "最小下单数量:%s", d.LotSize)
	}
	if d.TickSize.IsZero() {
		return errorx.Newf(errCode.INVALID_TICK_SIZE, "TickSize:%s", d.TickSize)
	}
	return nil
}

func (d DynamicSymbol) equal(other DynamicSymbol) bool {
	return d.UpLimitPercent.Equal(other.UpLimitPercent) &&
		d.DownLimitPercent.Equal(other.DownLimitPercent) &&
		d.MinQty.Equal(other.MinQty) &&
		d.TickSize.Equal(other.TickSize) &&
		d.LotSize.Equal(other.LotSize)
}

func (d DynamicSymbol) PrintMe() {
	dynamicLog.Log.GetLog().Infof("动态交易对信息:涨停百分比:%d,跌停百分比:%d,最小下单金额:%d,价格:[%d],数量:[%d]",
		d.UpLimitPercent, d.DownLimitPercent, d.MinQty, d.TickSize, d.LotSize)
}
