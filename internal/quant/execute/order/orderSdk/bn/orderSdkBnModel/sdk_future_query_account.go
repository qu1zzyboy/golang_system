package orderSdkBnModel

import (
	"strconv"
	"time"

	"github.com/hhh500/quantGoInfra/pkg/container/pool/byteBufPool"
	"github.com/hhh500/quantGoInfra/pkg/utils/convertx"
	"github.com/hhh500/quantGoInfra/pkg/utils/timeUtils"
	"github.com/hhh500/upbitBnServer/internal/utils/myCrypto"
)

type FutureQueryAccount struct {
}

// ParseRestReqFast 172.3 ns/op           184 B/op          5 allocs/op
func (api *FutureQueryAccount) ParseRestReqFast() *[]byte {
	orig := byteBufPool.AcquireBuffer(128)
	*orig = append(*orig, b_TIME_STAMP...)
	*orig = convertx.AppendValueToBytes(*orig, timeUtils.GetNowTimeUnixMilli())
	return orig
}

func (api *FutureQueryAccount) ParseWsReqFast(apiKey, method string, secretByte []byte) ([]byte, error) {
	param := make(map[string]any)
	ts := timeUtils.GetNowTimeUnixMilli()
	//统一逻辑
	param[p_API_KEY] = apiKey
	param[p_TIME_STAMP] = ts
	signRaw := buildQueryBytePool(128, param, querySortedKeyFast) //从池子中获取128位签名数据
	signRes := byteBufPool.AcquireBuffer(64)                      //从池子中获取64位
	defer byteBufPool.ReleaseBuffer(signRaw)                      //释放签名数据
	defer byteBufPool.ReleaseBuffer(signRes)                      //释放签名值
	if err := myCrypto.HmacSha256Fast(secretByte, *signRaw, signRes); err != nil {
		return nil, err
	}
	buf := make([]byte, 0, 512)
	buf = append(buf, `{"id":"`...)
	buf = append(buf, "605a6d20-6588-4cb9-afa0-b0ab087507ba"...)
	buf = append(buf, `","method":"v2/account.balance","params":{"apiKey":"`...)
	buf = append(buf, apiKey...)
	buf = append(buf, `","signature":"`...)
	buf = append(buf, (*signRes)...)
	buf = append(buf, `","timestamp":`...)
	buf = strconv.AppendInt(buf, ts, 10)
	buf = append(buf, `}}`...)
	return buf, nil
	// return buildWsReqFast(512, "605a6d20-6588-4cb9-afa0-b0ab087507ba", method, param, querySortedKeyFast, signRes), nil
}

func (api *FutureQueryAccount) ParseWsReqFastNoSign(id string) ([]byte, error) {
	buf := make([]byte, 0, 512)
	buf = append(buf, `{"id":"`...)
	buf = append(buf, id...)
	buf = append(buf, `","method":"v2/account.balance","params":{"timestamp":`...)
	buf = strconv.AppendInt(buf, time.Now().UnixMilli(), 10)
	buf = append(buf, `}}`...)
	return buf, nil
}

// NewFutureQueryAccount   rest查询订单 (USER_DATA)
func NewFutureQueryAccount() *FutureQueryAccount {
	return &FutureQueryAccount{}
}
