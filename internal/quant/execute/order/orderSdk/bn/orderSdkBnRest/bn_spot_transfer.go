package orderSdkBnRest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hhh500/quantGoInfra/infra/httpx"
	"github.com/hhh500/quantGoInfra/quant/exchanges/binance/bnConst"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderSdk/bn/orderSdkBnModel"
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
