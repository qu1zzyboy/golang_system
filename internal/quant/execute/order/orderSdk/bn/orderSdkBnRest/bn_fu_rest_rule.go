package orderSdkBnRest

import (
	"context"
	"fmt"
	"net/http"
	"upbitBnServer/internal/infra/httpx"
	"upbitBnServer/internal/quant/exchanges/binance/bnConst"
	"upbitBnServer/internal/quant/execute/order/orderSdk/bn/orderSdkBnModel"
)

// /fapi/v1/apiTradingStatus

var fuTradeRuleFullUrlByte = fmt.Appendf(nil, "%s/fapi/v1/symbolConfig?", bnConst.FUTURE_BASE_REST_URL)

// Do send request
func (s *FutureRest) DoTradeRule(ctx context.Context, api *orderSdkBnModel.FutureTradeingRuleSdk) ([]byte, error) {
	var urlByte []byte
	urlByte = append(urlByte, fuTradeRuleFullUrlByte...)
	r, err := s.addSignParamsGet(urlByte, api.ParseRestReq())
	if err != nil {
		return nil, err
	}
	r.Method = http.MethodGet
	body, err := httpx.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	return body, nil
}
