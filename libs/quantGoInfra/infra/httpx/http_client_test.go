package httpx

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/hhh500/quantGoInfra/quant/exchanges/binance/bnConst"
)

func TestHttpClient_Do(t *testing.T) {
	// 1. 解析 URL
	parsedURL, err := url.Parse(fmt.Sprintf("%s/fapi/v1/time", bnConst.FUTURE_BASE_REST_URL))
	if err != nil {
		t.Fatalf("URL parse failed: %v", err)
	}
	// 2. 构造请求
	req := &HttpRequest{
		Method:  http.MethodGet,
		URL:     parsedURL,
		Header:  http.Header{},
		Context: context.Background(),
	}
	// 3. 创建客户端并发送请求
	client := newHttpClient()
	respBody, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	t.Logf("==Binance Time API Response: %s", string(respBody))
}

// go test -v -run ^TestHttpClient_Do$
