package convertx

import (
	"fmt"
	"strconv"
)

func AppendValueToBytes(buf []byte, v any) []byte {
	switch val := v.(type) {
	case string:
		return append(buf, val...)
	case uint64:
		return strconv.AppendUint(buf, val, 10)
	case int:
		return strconv.AppendInt(buf, int64(val), 10)
	case int64:
		return strconv.AppendInt(buf, val, 10)
	case float64:
		return strconv.AppendFloat(buf, val, 'f', -1, 64)
	case fmt.Stringer:
		return append(buf, val.String()...)
	default:
		return append(buf, fmt.Sprintf("%v", val)...)
	}
}
