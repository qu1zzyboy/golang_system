package instanceEnum

type Type uint8

const (
	UNKNOWN              Type = iota //未知实例类型
	TEST                             //
	TO_UPBIT_LIST_BN                 //upbit上市实例
	TO_UPBIT_LIST_BYBIT              //bybit上市实例
	DOWNLOAD_ONLY_BN                 //bn下载服务
	CHECK_ALL_HEART_BEAT             //检查所有服务心跳实例
)

func (s Type) String() string {
	switch s {
	case TEST:
		return "TEST"
	case TO_UPBIT_LIST_BN:
		return "TO_UPBIT_LIST_BN"
	case TO_UPBIT_LIST_BYBIT:
		return "TO_UPBIT_LIST_BYBIT"
	case DOWNLOAD_ONLY_BN:
		return "DOWNLOAD_ONLY_BN"
	case CHECK_ALL_HEART_BEAT:
		return "CHECK_ALL_HEART_BEAT"
	default:
		return "ERROR"
	}
}
