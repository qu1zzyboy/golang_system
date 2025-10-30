package httpx

import (
	"io"
	"net/http"
	"time"

	"upbitBnServer/internal/infra/errorx"
	"upbitBnServer/internal/infra/errorx/errCode"
)

var (
	DefaultClient = newHttpClient()
	UrlParseErr   = errorx.New(errCode.HTTP_PARAM_ERROR, "URL解析错误")
)

type HttpClient struct {
	client *http.Client
}

func newHttpClient() *HttpClient {
	return &HttpClient{
		client: &http.Client{
			Timeout:   time.Second * 3,
			Transport: http.DefaultTransport, // 或自定义高性能 Transport
		},
	}
}

func (hc *HttpClient) Do(req *HttpRequest) ([]byte, error) {
	httpReq := &http.Request{
		Method: req.Method,
		URL:    req.URL,
		Header: req.Header,
	}
	if req.Body != nil {
		httpReq.Body = io.NopCloser(req.Body)
	}
	// httpReq = httpReq.WithContext(req.Context)
	resp, err := hc.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

//继续优化思路
//fasthttp和零拷贝读取响应体
