package myMap

import (
	"testing"
)

func TestMySyncMap(t *testing.T) {
	m := NewMySyncMap[string, int]()
	m.Store("a", 1)
	m.Store("b", 2)

	v, ok := m.Load("a")
	if !ok || v != 1 {
		t.Errorf("expect a=1, got %v", v)
	}

	m.Delete("a")
	if m.Length() != 1 {
		t.Errorf("expect length=1, got %d", m.Length())
	}
}
