package execute

import (
	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/infra/errorx"
	"upbitBnServer/internal/infra/errorx/errDefine"
	"upbitBnServer/pkg/utils/convertx"
)

// MyOrderMode 自定义订单模式,BUY_OPEN,SELL_OPEN,BUY_CLOSE,SELL_CLOSE
const (
	value_BUY_OPEN   = "BUY_OPEN"
	value_SELL_OPEN  = "SELL_OPEN"
	value_BUY_CLOSE  = "BUY_CLOSE"
	value_SELL_CLOSE = "SELL_CLOSE"
	ERROR_UPPER_CASE = "ERROR"
)

type MyOrderMode uint8

const (
	ORDER_BUY_OPEN   MyOrderMode = iota // 买入开多
	ORDER_SELL_OPEN                     // 卖出开空
	ORDER_BUY_CLOSE                     // 买入平空
	ORDER_SELL_CLOSE                    // 卖出平多
	ORDER_MODE_ERROR                    // ERROR
)

func (s MyOrderMode) Verify() *errorx.Error {
	switch s {
	case ORDER_BUY_OPEN, ORDER_SELL_OPEN, ORDER_BUY_CLOSE, ORDER_SELL_CLOSE:
		return nil
	default:
		return errDefine.EnumDefineError.WithMetadata(map[string]string{
			defineJson.Value:    convertx.ToString(s),
			defineJson.EnumType: "MyOrderMode",
		})
	}
}

func (s MyOrderMode) String() string {
	switch s {
	case ORDER_BUY_OPEN:
		return value_BUY_OPEN
	case ORDER_SELL_OPEN:
		return value_SELL_OPEN
	case ORDER_BUY_CLOSE:
		return value_BUY_CLOSE
	case ORDER_SELL_CLOSE:
		return value_SELL_CLOSE
	default:
		return ERROR_UPPER_CASE
	}
}

func (s MyOrderMode) GetOrderDirAndPositionDir() (MyOrderDir, MyPositionDir) {
	switch s {
	case ORDER_BUY_OPEN:
		return ORDER_BUY, POSITION_LONG
	case ORDER_SELL_OPEN:
		return ORDER_SELL, POSITION_SHORT
	case ORDER_BUY_CLOSE:
		return ORDER_BUY, POSITION_SHORT
	case ORDER_SELL_CLOSE:
		return ORDER_SELL, POSITION_LONG
	default:
		return ORDER_ERROR, POSITION_ERROR
	}
}

func (s MyOrderMode) IsBuy() bool {
	return s == ORDER_BUY_OPEN || s == ORDER_BUY_CLOSE
}

func (s MyOrderMode) IsOpen() bool {
	return s == ORDER_BUY_OPEN || s == ORDER_SELL_OPEN
}

func (s MyOrderMode) IsLong() bool {
	return s == ORDER_BUY_OPEN || s == ORDER_SELL_CLOSE
}

func GetOrderModeByEnum(orderDir MyOrderDir, posDir MyPositionDir) MyOrderMode {
	if orderDir == ORDER_BUY {
		if posDir == POSITION_LONG {
			return ORDER_BUY_OPEN
		} else {
			return ORDER_BUY_CLOSE
		}
	} else {
		if posDir == POSITION_LONG {
			return ORDER_SELL_CLOSE
		} else {
			return ORDER_SELL_OPEN
		}
	}
}

func GetOrderModeByString(s string) MyOrderMode {
	switch s {
	case value_BUY_OPEN:
		return ORDER_BUY_OPEN
	case value_SELL_OPEN:
		return ORDER_SELL_OPEN
	case value_BUY_CLOSE:
		return ORDER_BUY_CLOSE
	case value_SELL_CLOSE:
		return ORDER_SELL_CLOSE
	default:
		return ORDER_MODE_ERROR
	}
}

func GetOrderMode_(isOpen, isLong bool) MyOrderMode {
	if isOpen {
		if isLong {
			return ORDER_BUY_OPEN
		} else {
			return ORDER_SELL_OPEN
		}
	} else {
		if isLong {
			return ORDER_SELL_CLOSE
		} else {
			return ORDER_BUY_CLOSE
		}
	}
}

func GetIsOpen(isBuy, isLong bool) bool {
	if isBuy {
		if isLong {
			return true
		} else {
			return false
		}
	} else {
		if isLong {
			return false
		} else {
			return true
		}
	}
}
