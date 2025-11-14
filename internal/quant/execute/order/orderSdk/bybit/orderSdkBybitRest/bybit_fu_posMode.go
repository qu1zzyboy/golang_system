package orderSdkBybitRest

import (
	"fmt"
	"net/http"
	"strconv"
	"upbitBnServer/internal/infra/httpx"
	"upbitBnServer/internal/quant/exchanges/bybit/bybitConst"
)

var fuPosModeUrl = fmt.Sprintf("%s/v5/position/switch-mode", bybitConst.BASE_URL)

func (s *FutureRest) DoPosModeAll(mode uint8, coin string) ([]byte, error) {
	orig := make([]byte, 0, 128)
	orig = append(orig, `{"category":"linear","mode":"`...)
	orig = strconv.AppendUint(orig, uint64(mode), 10)
	orig = append(orig, `","coin":"`...)
	orig = append(orig, coin...)
	orig = append(orig, `"}`...)

	r, err := s.addSignPost(fuPosModeUrl, orig)
	if err != nil {
		return nil, err
	}
	r.Method = http.MethodPost
	return httpx.DefaultClient.Do(r)
}
