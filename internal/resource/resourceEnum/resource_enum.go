package resourceEnum

import (
	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/infra/errorx"
	"upbitBnServer/internal/infra/errorx/errCode"
	"upbitBnServer/internal/infra/errorx/errDefine"
	"upbitBnServer/pkg/utils/convertx"
)

type ResourceType uint8

// 行情资源
const (
	DELTA_DEPTH ResourceType = iota //全量深度数据
	LEVEL_DEPTH                     //深度数据
	ALL_DEPTH                       //全量深度数据
	BOOK_TICK                       //最优挂单数据
	KLINE                           //K线数据
	AGG_TRADE                       //聚合交易数据
	MARK_PRICE                      //标记价格数据
	FORCE_ORDER                     //强平订单数据
	PRICE_LIMIT                     //价格限制数据

	// 私有数据
	ORDER_WRITE  //订单ws连接
	PAYLOAD_READ //私有数据读取
)

func (s ResourceType) GetNotSupportError(flag string) error {
	return errorx.Newf(errCode.ENUM_NOT_SUPPORTED, "RESOURCE_TYPE_NOT_SUPPORT[%s] %s ", s.String(), flag)
}

func (s ResourceType) Verify() error {
	switch s {
	case DELTA_DEPTH, LEVEL_DEPTH, ALL_DEPTH, BOOK_TICK, KLINE, AGG_TRADE, MARK_PRICE, PRICE_LIMIT, ORDER_WRITE, FORCE_ORDER:
		return nil
	default:
		return errDefine.EnumDefineError.WithMetadata(map[string]string{
			defineJson.EnumType: "ResourceType",
			defineJson.Value:    convertx.ToString(s),
		})
	}
}

func (s ResourceType) String() string {
	switch s {
	case DELTA_DEPTH:
		return "DELTA_DEPTH"
	case LEVEL_DEPTH:
		return "LEVEL_DEPTH"
	case ALL_DEPTH:
		return "ALL_DEPTH"
	case BOOK_TICK:
		return "BOOK_TICKER"
	case KLINE:
		return "KLINE"
	case AGG_TRADE:
		return "AGG_TRADE"
	case FORCE_ORDER:
		return "FORCE_ORDER"
	case MARK_PRICE:
		return "MARK_PRICE"
	case PRICE_LIMIT:
		return "PRICE_LIMIT"
	case ORDER_WRITE:
		return "ORDER_WRITE"
	case PAYLOAD_READ:
		return "PRIVATE_READ"
	default:
		return "ERROR"
	}
}
