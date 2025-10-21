package execute

// MyPositionDir 自定义仓位方向
const (
	LONG_VALUE  = "LONG"
	SHORT_VALUE = "SHORT"
)

type MyPositionDir uint8

const (
	POSITION_LONG  MyPositionDir = iota // 多头
	POSITION_SHORT                      // 空头
	POSITION_ERROR                      // ERROR
)

func (s MyPositionDir) IsLong() bool {
	return s == POSITION_LONG
}

func (s MyPositionDir) String() string {
	switch s {
	case POSITION_LONG:
		return LONG_VALUE
	case POSITION_SHORT:
		return SHORT_VALUE
	default:
		return ERROR_UPPER_CASE
	}
}

func GetPositionDir(s string) MyPositionDir {
	switch s {
	case LONG_VALUE:
		return POSITION_LONG
	case SHORT_VALUE:
		return POSITION_SHORT
	default:
		return POSITION_ERROR
	}
}

func GetPositionDirValue(isLong bool) string {
	if isLong {
		return LONG_VALUE
	}
	return SHORT_VALUE
}

func GetIsLong(s string) bool {
	return s == LONG_VALUE
}
