package orderSdkBnRest

import (
	"context"
	"fmt"
	"net/http"

	"upbitBnServer/internal/infra/httpx"
	"upbitBnServer/internal/quant/exchanges/binance/bnConst"
	"upbitBnServer/internal/quant/execute/order/orderSdk/bn/orderSdkBnModel"
)

var (
	spotTransferFullUrlByte = fmt.Appendf(nil, "%s/sapi/v1/sub-account/universalTransfer?", bnConst.SPOT_BASE_REST_URL)
)

func (s *SpotRest) DoTransfer(ctx context.Context, api *orderSdkBnModel.UniversalTransferSdk) (string, string, error) {
	var urlByte []byte
	urlByte = append(urlByte, spotTransferFullUrlByte...)
	reqByte := api.ParseRestReq()
	r, err := s.addSignParamsFast(urlByte, reqByte)
	if err != nil {
		return "", "", err
	}
	r.Method = http.MethodPost
	body, err := httpx.DefaultClient.Do(r)
	if err != nil {
		return "", "", err
	}
	return string(body), string(reqByte), nil
}
