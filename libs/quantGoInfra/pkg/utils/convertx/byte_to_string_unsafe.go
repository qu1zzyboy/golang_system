package convertx

import "unsafe"

func BytesToUnsafeString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
