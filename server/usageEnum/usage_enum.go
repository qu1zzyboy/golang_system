package usageEnum

type Type uint8

const (
	UNKNOWN             Type = iota //未知订单状态
	NEWS_DRIVE_PRE                  //UpBit待上市预挂单
	NEWS_DRIVE_MAIN                 //UpBit待上市循环挂单
	CANCEL_AND_TRANSFER             //UpBit待上市循环撤单并划转

)

func (s Type) String() string {
	switch s {
	case NEWS_DRIVE_PRE:
		return "NEWS_DRIVE_PRE"
	case NEWS_DRIVE_MAIN:
		return "NEWS_DRIVE_MAIN"
	case CANCEL_AND_TRANSFER:
		return "CANCEL_AND_TRANSFER"
	default:
		return "ERROR"
	}
}
