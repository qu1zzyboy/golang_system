package orderSdkBnRest

import (
	"fmt"
	"net/http"

	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/infra/errorx"
	"upbitBnServer/internal/infra/errorx/errCode"
	"upbitBnServer/internal/infra/httpx"
	"upbitBnServer/internal/quant/exchanges/binance/bnConst"

	"github.com/tidwall/gjson"
)

func (s *FutureRest) DoListenKey() (string, error) {
	r, err := s.addNoSignParamsFast(fmt.Sprintf("%s/fapi/v1/listenKey", bnConst.FUTURE_BASE_REST_URL))
	if err != nil {
		return "", err
	}
	r.Method = http.MethodPost
	body, err := httpx.DefaultClient.Do(r)
	if err != nil {
		return "", err
	}
	jsonStr := string(body)
	listenKey := gjson.Get(jsonStr, "listenKey").String()
	if listenKey != "" {
		return listenKey, nil
	}
	return "", errorx.New(errCode.HTTP_DO_ERROR, "bn获取listenKey失败").WithMetadata(map[string]string{
		defineJson.FullUrl: r.URL.String(),
		defineJson.RawJson: jsonStr,
	})
}

func (s *FutureRest) DelayListenKey(listenKey string) error {
	r, err := s.addNoSignParamsFast(fmt.Sprintf("%s/fapi/v1/listenKey?listenKey=%s", bnConst.FUTURE_BASE_REST_URL, listenKey))
	if err != nil {
		return err
	}
	r.Method = http.MethodPut
	_, err = httpx.DefaultClient.Do(r)
	return err
}
