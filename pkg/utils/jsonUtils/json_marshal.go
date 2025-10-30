package jsonUtils

import (
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// jsoniter：性能比标准库高 20%-30%

func MarshalStructToByteArray(v any) ([]byte, error) {
	return json.Marshal(v)
}

func MarshalStructToString(v any) (string, error) {
	result, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

func UnmarshalFromByteArray(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func UnmarshalFromString(str string, v interface{}) error {
	return json.UnmarshalFromString(str, v)
}
