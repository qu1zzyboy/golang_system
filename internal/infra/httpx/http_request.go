package httpx

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

// HttpRequest 封装一个 HTTP 请求所需的所有参数
type HttpRequest struct {
	Method  string
	URL     *url.URL
	Header  http.Header
	Body    io.Reader
	Context context.Context
}

func GetCommonHttpRequest(urlStr string) (*HttpRequest, error) {
	// 构建请求头 Header
	httpHeader := make(http.Header)
	httpHeader.Set("Content-Type", "application/json")
	// URL 解析
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	return &HttpRequest{
		URL:    parsedURL,
		Header: httpHeader,
		Body:   http.NoBody,
	}, nil
}
