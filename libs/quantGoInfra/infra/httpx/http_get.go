package httpx

import (
	"net/http"

	"github.com/hhh500/quantGoInfra/define/defineJson"
	"github.com/hhh500/quantGoInfra/infra/errorx/errDefine"
)

func Get(fullUrl string) ([]byte, error) {
	httpReq, err := GetCommonHttpRequest(fullUrl)
	if err != nil {
		return nil, errDefine.HttpParamError.WithMetadata(map[string]string{defineJson.FullUrl: fullUrl}).WithCause(err)
	}
	httpReq.Method = http.MethodGet
	data, err := DefaultClient.Do(httpReq)
	if err != nil {
		return nil, errDefine.HttpDoError.WithMetadata(map[string]string{defineJson.FullUrl: fullUrl}).WithCause(err)
	}
	return data, nil
}
