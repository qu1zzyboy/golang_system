package debugx

import (
	"testing"

	"github.com/hhh500/quantGoInfra/infra/observe/log/staticLog"
)

func TestXxx(t *testing.T) {
	staticLog.Log.Info(GetCaller(1))
}
