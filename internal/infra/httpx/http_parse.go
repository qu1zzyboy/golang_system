package httpx

import (
	"upbitBnServer/pkg/utils/convertx"
)

func BuildQueryByte(preSize int, params map[string]interface{}, keySorted []string) []byte {
	if len(params) == 0 {
		return nil
	}
	b := make([]byte, 0, preSize)
	for i, k := range keySorted {
		if val, ok := params[k]; ok {
			if i > 0 {
				b = append(b, '&')
			}
			b = append(b, k...)
			b = append(b, '=')
			b = convertx.AppendValueToBytes(b, val)
		}
	}
	return b
}
