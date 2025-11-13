package orderSdkBybitRest

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"upbitBnServer/internal/infra/httpx"
	"upbitBnServer/internal/quant/exchanges/bybit/bybitConst"
	"upbitBnServer/internal/utils/myCrypto"
	"upbitBnServer/pkg/container/pool/byteBufPool"
	"upbitBnServer/pkg/utils/convertx"
	"upbitBnServer/pkg/utils/timeUtils"
)

var fuLeverageUrl = fmt.Sprintf("%s/v5/position/set-leverage", bybitConst.BASE_URL)

func (s *FutureRest) DoLeverage(leverage uint8, symbolName string) ([]byte, error) {
	orig := make([]byte, 0, 128)
	orig = append(orig, `{"category":"linear","buyLeverage":"`...)
	orig = strconv.AppendUint(orig, uint64(leverage), 10)
	orig = append(orig, `","sellLeverage":"`...)
	orig = strconv.AppendUint(orig, uint64(leverage), 10)
	orig = append(orig, `","symbol":"`...)
	orig = append(orig, symbolName...)
	orig = append(orig, `"}`...)

	r, err := s.addSignDoLeverage(fuLeverageUrl, orig)
	if err != nil {
		return nil, err
	}
	r.Method = http.MethodPost
	return httpx.DefaultClient.Do(r)
}

func (s *FutureRest) addSignDoLeverage(urlStr string, param []byte) (*httpx.HttpRequest, error) {
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
