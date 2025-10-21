package execute

type OrderStatus uint8

const (
	UNKNOWN_ORDER_STATUS = iota //未知状态
	NEW                         //新订单已被引擎接受。
	PARTIALLY_FILLED            //部分成交
	FILLED                      //订单已完全成交。
	CANCELED                    //订单被用户取消。
	REJECTED                    //新订单被拒绝 这信息只会在撤消挂单再下单中发生，下新订单被拒绝但撤消挂单请求成功
	EXPIRED                     //订单已根据 ArrTime In Force 参数的规则取消

)

func ParseBnOrderStatus(status string) OrderStatus {
	switch status {
	case "NEW":
		return NEW
	case "PARTIALLY_FILLED":
		return PARTIALLY_FILLED
	case "FILLED":
		return FILLED
	case "CANCELED":
		return CANCELED
	case "REJECTED":
		return REJECTED
	case "EXPIRED":
		return EXPIRED
	default:
		return UNKNOWN_ORDER_STATUS // 未知状态
	}
}

func ParseByBitOrderStatus(status string) OrderStatus {
	switch status {
	case "New":
		return NEW
	case "PartiallyFilled":
		return PARTIALLY_FILLED
	case "Filled":
		return FILLED
	case "Cancelled":
		return CANCELED
	case "Rejected":
		return REJECTED
	default:
		return UNKNOWN_ORDER_STATUS // 未知状态
	}
}

func IsOrderOnLine(status OrderStatus) bool {
	switch status {
	case NEW, PARTIALLY_FILLED:
		return true
	default:
		return false
	}
}
