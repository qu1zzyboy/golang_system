package orderBelongEnum

type Type uint8

const (
	UNKNOWN            Type = iota //未知订单状态
	TO_UPBIT_LIST_PRE              //UpBit待上市预挂单
	TO_UPBIT_LIST_LOOP             //UpBit待上市循环挂单
	TO_UPBIT_LIST_MONITOR
	TO_UPBIT_LIST_LOOP_CANCEL_TRANSFER //UpBit待上市循环撤单并划转
)
