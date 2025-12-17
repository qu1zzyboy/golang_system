package instanceEnum

import "upbitBnServer/internal/infra/systemx/instance/instanceDefine"

const (
	DRIVER_LIST_BN instanceDefine.Type = iota
	TO_UPBIT_ON_LIST
	TO_UPBIT_DOWN_LIST
	CHECK_HEART_BEAT
	PRINT_ALL_INSTANCE
	TEST
)

func String(s instanceDefine.Type) string {
	switch s {
	case TEST:
		return "TEST"
	case DRIVER_LIST_BN:
		return "DRIVER_LIST_BN"
	default:
		return "ERROR"
	}
}
