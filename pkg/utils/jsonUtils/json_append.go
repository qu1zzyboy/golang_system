package jsonUtils

import "strconv"

func InjectReceiveAt(jsonBytes []byte, ts int64) []byte {
	// 检查 JSON 是否是对象形式(以 { 开始,以 } 结束)
	if len(jsonBytes) < 2 || jsonBytes[0] != '{' || jsonBytes[len(jsonBytes)-1] != '}' {
		return jsonBytes
	}
	// 去掉末尾的 '}'
	base := jsonBytes[:len(jsonBytes)-1]
	// 构造 ,"receiveAt":<ts>
	buf := []byte(`,"receiveAt":`)
	buf = strconv.AppendInt(buf, ts, 10)
	// 重新拼接并加回 }
	return append(append(base, buf...), '}')
}

func InjectReceiveDiff(jsonBytes []byte, ts int64) []byte {
	// 检查 JSON 是否是对象形式(以 { 开始,以 } 结束)
	if len(jsonBytes) < 2 || jsonBytes[0] != '{' || jsonBytes[len(jsonBytes)-1] != '}' {
		return jsonBytes
	}
	// 去掉末尾的 '}'
	base := jsonBytes[:len(jsonBytes)-1]
	// 构造 ,"receiveAt":<ts>
	buf := []byte(`,"receiveDiff":`)
	buf = strconv.AppendInt(buf, ts, 10)
	// 重新拼接并加回 }
	return append(append(base, buf...), '}')
}
