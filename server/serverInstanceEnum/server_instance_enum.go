package serverInstanceEnum

type Type uint32

const (
	UNKNOWN              Type = iota //为止实例类型
	TO_UPBIT_LIST_BN                 //upbit上市实例
	TO_UPBIT_LIST_BYBIT              //bybit上市实例
	CHECK_ALL_HEART_BEAT             //检查所有服务心跳实例
)
