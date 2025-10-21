package execute

import (
	"github.com/hhh500/quantGoInfra/define/defineJson"
	"github.com/hhh500/quantGoInfra/infra/errorx"
	"github.com/hhh500/quantGoInfra/infra/errorx/errDefine"
	"github.com/hhh500/quantGoInfra/pkg/utils/convertx"
)

// MyOrderPlaceType 自定义挂单类型
type MyOrderPlaceType uint8

const (
	oRDER_TYPE_LIMIT_VALUE       = "LIMIT"
	oRDER_TYPE_POST_ONLY_VALUE   = "LIMIT_MAKER"
	oRDER_TYPE_IOC_VALUE         = "IOC"
	oRDER_TYPE_GTD_VALUE         = "GTD"
	oRDER_TYPE_STOP_MARKET_VALUE = "STOP_MARKET"
)

const (
	ORDER_TYPE_LIMIT       MyOrderPlaceType = iota // 限价单
	ORDER_TYPE_POST_ONLY                           // 限价只做maker单
	ORDER_TYPE_MARKET                              // 市价单
	ORDER_TYPE_IOC                                 // IOC,立即成交或取消
	ORDER_TYPE_GTD                                 // GTD,超时取消
	ORDER_TYPE_STOP_MARKET                         // 止损市价单
	ORDER_TYPE_ERROR                               // ERROR
)

func GetMyOrderType(s string) MyOrderPlaceType {
	switch s {
	case oRDER_TYPE_LIMIT_VALUE:
		return ORDER_TYPE_LIMIT
	case oRDER_TYPE_POST_ONLY_VALUE:
		return ORDER_TYPE_POST_ONLY
	case oRDER_TYPE_IOC_VALUE:
		return ORDER_TYPE_IOC
	case oRDER_TYPE_GTD_VALUE:
		return ORDER_TYPE_GTD
	case oRDER_TYPE_STOP_MARKET_VALUE:
		return ORDER_TYPE_STOP_MARKET
	default:
		return ORDER_TYPE_ERROR
	}
}

func (s MyOrderPlaceType) Verify() *errorx.Error {
	if s >= ORDER_TYPE_ERROR {
		return errDefine.EnumDefineError.WithMetadata(map[string]string{
			defineJson.Value:    convertx.ToString(s),
			defineJson.EnumType: "MyOrderPlaceType",
		})
	}
	return nil
}

func (s MyOrderPlaceType) String() string {
	switch s {
	case ORDER_TYPE_LIMIT:
		return oRDER_TYPE_LIMIT_VALUE
	case ORDER_TYPE_POST_ONLY:
		return oRDER_TYPE_POST_ONLY_VALUE
	case ORDER_TYPE_IOC:
		return oRDER_TYPE_IOC_VALUE
	case ORDER_TYPE_GTD:
		return oRDER_TYPE_GTD_VALUE
	case ORDER_TYPE_STOP_MARKET:
		return oRDER_TYPE_STOP_MARKET_VALUE
	default:
		return ERROR_UPPER_CASE
	}
}
