package aggTrade

import (
	"fmt"
	"strings"
)

var (
	BN    bn
	BYBIT bybit
)

type bn struct {
}
type bybit struct {
}

func (s bn) GetFuSubParam(symbol string) string {
	return fmt.Sprintf("%s@aggTrade", strings.ToLower(symbol))
}

func (s bybit) GetFuSubParam(symbol string) string {
	return fmt.Sprintf("publicTrade.%s", strings.ToUpper(symbol))
}
