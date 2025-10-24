package convertx

import (
	"fmt"
	"strconv"
)

const (
	TRUE_LOWER_CASE  = "true"
	FALSE_LOWER_CASE = "false"
)

func ToString(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case uint8:
		return strconv.Itoa(int(v))
	case uint16:
		return strconv.Itoa(int(v))
	case uint32:
		return strconv.Itoa(int(v))
	case uint64:
		return strconv.FormatUint(v, 10)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case bool:
		if v {
			return TRUE_LOWER_CASE
		}
		return FALSE_LOWER_CASE
	default:
		return fmt.Sprintf("%v", val)
	}
}
