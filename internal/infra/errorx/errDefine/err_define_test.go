package errDefine

import "testing"

func TestXxx(t *testing.T) {
	t.Log(PointerNil.WithMetadata(map[string]string{"key": "value"}))
}
