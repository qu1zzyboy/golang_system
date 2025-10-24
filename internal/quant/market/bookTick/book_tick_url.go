package bookTick

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
	return fmt.Sprintf("%s@bookTicker", strings.ToLower(symbol))
}

func (s bybit) GetFuSubParam(symbol string) string {
	return fmt.Sprintf("orderbook.1.%s", strings.ToUpper(symbol))
}
