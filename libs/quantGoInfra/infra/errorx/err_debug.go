package errorx

import (
	"fmt"

	"github.com/hhh500/quantGoInfra/infra/debugx"
)

func PanicWithCaller(msg string) {
	caller := debugx.GetCaller(2)
	panic(fmt.Sprintf("%s %s", caller, msg))
}
