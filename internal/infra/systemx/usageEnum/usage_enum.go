package usageEnum

type Type uint8

const (
	UNKNOWN                  Type = iota //未知订单状态
	TO_UPBIT_PRE                         //UpBit待上市预挂单
	TO_UPBIT_MONITOR                     //UpBit待上市循环挂单探测
	TO_UPBIT_MAIN                        //UpBit待上市循环挂单
	TO_UPBIT_CANCEL_TRANSFER             //UpBit待上市循环撤单并划转
)

func (s Type) String() string {
	switch s {
	default:
		return "ERROR"
	}
}
