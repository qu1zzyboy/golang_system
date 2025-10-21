package orderSdkBnRest

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/hhh500/quantGoInfra/define/defineJson"
	"github.com/hhh500/quantGoInfra/infra/httpx"
	"github.com/hhh500/quantGoInfra/pkg/container/pool/byteBufPool"
	"github.com/hhh500/quantGoInfra/quant/exchanges/binance/bnConst"
	"github.com/hhh500/upbitBnServer/internal/utils/myCrypto"
)

const (
	key_X_MBX_APIKEY = "X-MBX-APIKEY"
	key_SIGNATURE    = "signature"
)

var (
	futureBatchFullUrl = fmt.Sprintf("%s/fapi/v1/batchOrders?", bnConst.FUTURE_BASE_REST_URL)
	futureFullUrlBytes = fmt.Appendf(nil, "%s/fapi/v1/order?", bnConst.FUTURE_BASE_REST_URL)
	b_SIGNATURE_Equal  = []byte("&signature=")
)

type FutureRest struct {
	apiKey     string
	secretKey  string
	secretByte []byte
	httpHeader http.Header
}

func NewFutureRest(apiKey, secretKey string) *FutureRest {
	s := &FutureRest{
		apiKey:    apiKey,
		secretKey: secretKey,
	}
	s.secretByte = []byte(secretKey)
	s.httpHeader = make(http.Header)
	s.httpHeader.Set("Content-Type", "application/json")
	s.httpHeader.Set(key_X_MBX_APIKEY, apiKey)
	return s
}

func (s *FutureRest) addSignParamsFast(urlByte, param *[]byte) (*httpx.HttpRequest, error) {
	defer byteBufPool.ReleaseBuffer(urlByte)
	defer byteBufPool.ReleaseBuffer(param)
	//1、 生成签名
	signByte := byteBufPool.AcquireBuffer(64)
	defer byteBufPool.ReleaseBuffer(signByte)
	if err := myCrypto.HmacSha256Fast(s.secretByte, *param, signByte); err != nil {
		return nil, err
	}
	*param = append(*param, b_SIGNATURE_Equal...)
	*param = append(*param, *signByte...)
	*urlByte = append(*urlByte, *param...)
	// URL 解析
	parsedURL, err := url.Parse(string(*urlByte))
	if err != nil {
		return nil, httpx.UrlParseErr.WithCause(err).WithMetadata(map[string]string{defineJson.FullUrl: string(*urlByte)})
	}
	return &httpx.HttpRequest{
		URL:    parsedURL,
		Header: s.httpHeader.Clone(),
		Body:   http.NoBody,
	}, nil
}

func (s *FutureRest) addNoSignParamsFast(fullURL string) (*httpx.HttpRequest, error) {
	// 构建请求头 Header
	header := s.httpHeader.Clone()
	// URL 解析
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return nil, err
	}
	return &httpx.HttpRequest{
		URL:    parsedURL,
		Header: header,
		Body:   http.NoBody,
	}, nil
}

func (s *FutureRest) addCommonParams(queryString, fullURL string) (*httpx.HttpRequest, error) {
	// 构建请求头 Header
	header := s.httpHeader.Clone()
	// 生成签名
	signature, err := myCrypto.Sha256(s.secretKey, queryString)
	if err != nil {
		return nil, err
	}
	queryString += "&" + key_SIGNATURE + "=" + *signature
	fullURL += "?" + queryString
	// URL 解析
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return nil, err
	}
	return &httpx.HttpRequest{
		URL:    parsedURL,
		Header: header,
		Body:   http.NoBody,
	}, nil
}
