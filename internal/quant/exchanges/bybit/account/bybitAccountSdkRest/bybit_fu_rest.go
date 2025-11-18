package bybitAccountSdkRest

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"upbitBnServer/internal/infra/httpx"
	"upbitBnServer/internal/quant/exchanges/bybit/bybitConst"
	"upbitBnServer/internal/utils/myCrypto"
	"upbitBnServer/pkg/container/pool/byteBufPool"
	"upbitBnServer/pkg/utils/convertx"
	"upbitBnServer/pkg/utils/timeUtils"
)

const (
	p_API_KEY            = "X-BAPI-API-KEY"
	p_SIGN_TYPE_KEY      = "X-BAPI-SIGN-TYPE"
	p_TIME_STAMP_KEY     = "X-BAPI-TIMESTAMP"
	p_SIGNATURE_KEY      = "X-BAPI-SIGN"
	p_RECEIVE_WINDOW_KEY = "X-BAPI-RECV-WINDOW"
)

var (
	b_RECEIVE_WINDOW_   = []byte("5000")
	futureFullPlaceStr  = fmt.Sprintf("%s/v5/order/create", bybitConst.BASE_URL)
	futureFullQueryByte = []byte(fmt.Sprintf("%s/v5/order/realtime", bybitConst.BASE_URL))
	futureFullCancelStr = fmt.Sprintf("%s/v5/order/cancel", bybitConst.BASE_URL)
	futureFullModifyStr = fmt.Sprintf("%s/v5/order/amend", bybitConst.BASE_URL)
)

type FutureRest struct {
	apiKey     string
	secretKey  string
	apiKeyByte []byte
	secretByte []byte
	httpHeader http.Header
}

func NewFutureRest(apiKey, secretKey string) *FutureRest {
	s := &FutureRest{
		apiKey:     apiKey,
		secretKey:  secretKey,
		apiKeyByte: []byte(apiKey),
		secretByte: []byte(secretKey),
	}
	s.httpHeader = make(http.Header)
	s.httpHeader.Set(p_SIGN_TYPE_KEY, "2")
	s.httpHeader.Set("Content-Type", "application/json")
	s.httpHeader.Set(p_API_KEY, apiKey)

	return s
}

func (s *FutureRest) addSignParamsFast_Body(urlStr string, param []byte) (*httpx.HttpRequest, error) {
	// 构建请求头 Header
	timeStamp := timeUtils.GetNowTimeUnixMilli()
	header := s.httpHeader.Clone()
	header.Set(p_TIME_STAMP_KEY, convertx.GetTimeStampMilliStrBy(timeStamp))
	header.Set(p_RECEIVE_WINDOW_KEY, "5000")
	// 生成签名
	signData := byteBufPool.AcquireBuffer(byteBufPool.SIZE_256)
	defer byteBufPool.ReleaseBuffer(signData)
	*signData = convertx.AppendValueToBytes(*signData, timeStamp) //添加时间戳
	*signData = append(*signData, s.apiKeyByte...)                //添加apiKey
	*signData = append(*signData, b_RECEIVE_WINDOW_...)           //添加接收窗口
	*signData = append(*signData, param...)                       //添加查询参数
	//得到签名结果
	signResp := byteBufPool.AcquireBuffer(byteBufPool.SIZE_64)
	defer byteBufPool.ReleaseBuffer(signResp)
	err := myCrypto.HmacSha256Fast(s.secretByte, *signData, signResp)
	if err != nil {
		return nil, err
	}
	header.Set(p_SIGNATURE_KEY, string(*signResp))
	// URL 解析
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	return &httpx.HttpRequest{
		URL:    parsedURL,
		Header: header,
		Body:   bytes.NewReader(param),
	}, nil
}

func (s *FutureRest) addSignParamsFast_Query(urlByte []byte, param *[]byte) (*httpx.HttpRequest, error) {
	// 构建请求头 Header
	timeStamp := timeUtils.GetNowTimeUnixMilli()
	header := s.httpHeader.Clone()
	header.Set(p_TIME_STAMP_KEY, convertx.GetTimeStampMilliStrBy(timeStamp))
	header.Set(p_RECEIVE_WINDOW_KEY, "5000")
	// 生成签名数据
	signData := byteBufPool.AcquireBuffer(byteBufPool.SIZE_128)
	*signData = convertx.AppendValueToBytes(*signData, timeStamp) //添加时间戳
	*signData = append(*signData, s.apiKeyByte...)                //添加apiKey
	*signData = append(*signData, b_RECEIVE_WINDOW_...)           //添加接收窗口
	*signData = append(*signData, *param...)                      //添加查询参数
	// myLog.LogDir.Info("Bybit sign: ", string(*signData))
	//得到签名结果
	signResp := byteBufPool.AcquireBuffer(byteBufPool.SIZE_64)
	defer func() {
		byteBufPool.ReleaseBuffer(param)
		byteBufPool.ReleaseBuffer(signData)
		byteBufPool.ReleaseBuffer(signResp)
	}()
	err := myCrypto.HmacSha256Fast(s.secretByte, *signData, signResp)
	if err != nil {
		return nil, err
	}
	// myLog.LogDir.Infof("signature: %s", string(*signResp))
	header.Set(p_SIGNATURE_KEY, string(*signResp))
	// URL 解析
	urlByte = append(urlByte, '?')
	urlByte = append(urlByte, *param...)
	parsedURL, err := url.Parse(string(urlByte))
	if err != nil {
		return nil, err
	}
	return &httpx.HttpRequest{
		URL:    parsedURL,
		Header: header,
		Body:   http.NoBody,
	}, nil
}
