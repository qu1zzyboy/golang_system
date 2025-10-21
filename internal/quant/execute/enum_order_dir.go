package execute

// MyOrderDir 自定义订单方向
type MyOrderDir uint8

const (
	ORDER_BUY_VALUE  = "BUY"
	ORDER_SELL_VALUE = "SELL"
)

const (
	ORDER_BUY   MyOrderDir = iota // 买入
	ORDER_SELL                    // 卖出
	ORDER_ERROR                   // ERROR
)

func (s MyOrderDir) IsBuy() bool {
	return s == ORDER_BUY
}

func GetMyOrderDir(s string) MyOrderDir {
	switch s {
	case ORDER_BUY_VALUE:
		return ORDER_BUY
	case ORDER_SELL_VALUE:
		return ORDER_SELL
	default:
		return ORDER_ERROR
	}
}

func (s MyOrderDir) String() string {
	switch s {
	case ORDER_BUY:
		return ORDER_BUY_VALUE
	case ORDER_SELL:
		return ORDER_SELL_VALUE
	default:
		return ERROR_UPPER_CASE
	}
}
