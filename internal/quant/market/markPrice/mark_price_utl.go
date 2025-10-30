package markPrice

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
	return fmt.Sprintf("%s@markPrice@1s", strings.ToLower(symbol))
}

func (s bybit) GetFuSubParam(symbol string) string {
	return fmt.Sprintf("tickers.%s", strings.ToUpper(symbol))
}
