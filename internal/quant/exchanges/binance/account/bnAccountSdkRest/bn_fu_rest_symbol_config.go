package bnAccountSdkRest

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"upbitBnServer/internal/quant/exchanges/binance/order/bnOrderSdkModel"

	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/infra/httpx"
	"upbitBnServer/internal/quant/exchanges/binance/bnConst"
	"upbitBnServer/internal/utils/myCrypto"
	"upbitBnServer/pkg/container/pool/byteBufPool"
)

var fuSymbolConfigFullUrlByte = fmt.Appendf(nil, "%s/fapi/v1/symbolConfig?", bnConst.FUTURE_BASE_REST_URL)

// Do send request
func (s *FutureRest) DoSymbolConfig(ctx context.Context, api *bnOrderSdkModel.FutureSymbolConfigSdk) ([]byte, error) {
	var urlByte []byte
	urlByte = append(urlByte, fuSymbolConfigFullUrlByte...)
	r, err := s.addSignParamsSymbolConfig(urlByte, api.ParseRestReq())
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

func (s *FutureRest) addSignParamsSymbolConfig(urlByte, param []byte) (*httpx.HttpRequest, error) {
	//1、 生成签名
	signByte := byteBufPool.AcquireBuffer(64)
	defer byteBufPool.ReleaseBuffer(signByte)
	if err := myCrypto.HmacSha256Fast(s.secretByte, param, signByte); err != nil {
		return nil, err
	}
	param = append(param, b_SIGNATURE_Equal...)
	param = append(param, *signByte...)
	urlByte = append(urlByte, param...)
	// URL 解析
	parsedURL, err := url.Parse(string(urlByte))
	if err != nil {
		return nil, httpx.UrlParseErr.WithCause(err).WithMetadata(map[string]string{defineJson.FullUrl: string(urlByte)})
	}
	return &httpx.HttpRequest{
		URL:    parsedURL,
		Header: s.httpHeader.Clone(),
		Body:   http.NoBody,
	}, nil
}
