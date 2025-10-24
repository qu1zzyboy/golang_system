package debugx

import (
	"testing"

	"upbitBnServer/internal/infra/observe/log/staticLog"
)

func TestXxx(t *testing.T) {
	staticLog.Log.Info(GetCaller(1))
}
