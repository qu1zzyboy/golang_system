package dynamicLog

import "testing"

func TestDynamicLogger(t *testing.T) {
	Error.GetLog().Info("This is a test log message")
	Error.GetLog().Error("This is a test error message")
}
