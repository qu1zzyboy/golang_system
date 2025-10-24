package orderSdkBnWsSign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strconv"
	"time"

	"upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"upbitBnServer/internal/quant/execute/order/wsRequestCache"
	"upbitBnServer/pkg/utils/convertx"
	"upbitBnServer/pkg/utils/jsonUtils"
)

// 生成 HMAC-SHA256 签名
func (s *FutureClient) hashing(query string) string {
	mac := hmac.New(sha256.New, s.secretByte)
	mac.Write([]byte(query))
	return hex.EncodeToString(mac.Sum(nil))
}

// 构建 account.balance 请求体
func (s *FutureClient) accountPositionInfo(reqId string) map[string]interface{} {
	timeStamp := time.Now().UnixMilli()
	// payload 用于签名
	values := url.Values{}
	values.Add("apiKey", s.apiKey)
	values.Add("timestamp", strconv.FormatInt(timeStamp, 10))

	params := map[string]any{
		"apiKey":    s.apiKey,
		"signature": s.hashing(values.Encode()),
		"timestamp": timeStamp,
	}
	return map[string]any{
		"id":           reqId,
		"method":       "v2/account.balance",
		"toUpbitParam": params,
	}
}

func (s *FutureClient) QueryAccount(reqFrom orderBelongEnum.Type) error {
	reqId := "qab" + convertx.GetNowTimeStampMilliStr()
	msg := s.accountPositionInfo(reqId)
	rawData, err := jsonUtils.MarshalStructToByteArray(msg)
	if rawData == nil || err != nil {
		return err
	}
	if err = s.conn.WriteAsync(rawData); err != nil {
		return err
	}
	wsRequestCache.GetCache().StoreMeta(reqId, &wsRequestCache.WsRequestMeta{
		Json:    string(rawData),
		ReqType: wsRequestCache.QUERY_ACCOUNT_BALANCE,
		ReqFrom: reqFrom,
	})
	return err
}
