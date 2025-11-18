package bnAccountSdkRest

import (
	"fmt"
	"net/http"
	"net/url"

	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/infra/httpx"
	"upbitBnServer/internal/quant/exchanges/binance/bnConst"
	"upbitBnServer/internal/utils/myCrypto"
	"upbitBnServer/pkg/container/pool/byteBufPool"
)

var (
	spotFullUrl = fmt.Sprintf("%s/api/v3/order?", bnConst.SPOT_BASE_REST_URL)
)

type SpotRest struct {
	apiKey     string
	secretKey  string
	secretByte []byte
	httpHeader http.Header
}

func NewSpotRest(apiKey, secretKey string) *SpotRest {
	s := &SpotRest{
		apiKey:    apiKey,
		secretKey: secretKey,
	}
	s.secretByte = []byte(secretKey)
	s.httpHeader = make(http.Header)
	s.httpHeader.Set("Content-Type", "application/json")
	s.httpHeader.Set(key_X_MBX_APIKEY, apiKey)
	return s
}

func (s *SpotRest) addCommonParamsFast(queryByte []byte, fullURL string) (*httpx.HttpRequest, error) {
	// 构建请求头 Header
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")
	headers.Set(key_X_MBX_APIKEY, s.apiKey)
	// 生成签名
	signature, err := myCrypto.HmacSha256(s.secretByte, queryByte)
	if err != nil {
		return nil, err
	}
	queryByte = append(queryByte, '&')
	queryByte = append(queryByte, key_SIGNATURE...)
	queryByte = append(queryByte, '=')
	queryString := string(queryByte)
	queryString += signature
	fullURL += "?" + queryString
	// URL 解析
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return nil, httpx.UrlParseErr.WithCause(err).WithMetadata(map[string]string{defineJson.FullUrl: fullURL})
	}
	return &httpx.HttpRequest{
		URL:    parsedURL,
		Header: headers,
		Body:   http.NoBody,
	}, nil
}

func (s *SpotRest) addSignParamsFast(urlByte, param []byte) (*httpx.HttpRequest, error) {
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
