package errorx

import (
	"fmt"

	"upbitBnServer/internal/infra/debugx"
)

func PanicWithCaller(msg string) {
	caller := debugx.GetCaller(2)
	panic(fmt.Sprintf("%s %s", caller, msg))
}
