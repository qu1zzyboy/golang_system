package orderSdkBnModel

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

type FutureLoginSdk struct {
	Id string
}

func (api *FutureLoginSdk) Id_(id string) *FutureLoginSdk {
	api.Id = id
	return api
}

var loginSortedKeyFast = []string{p_API_KEY, p_TIME_STAMP}

func (api *FutureLoginSdk) ParseWsReqFast(apiKey, method string, secretByte []byte) ([]byte, error) {
	// // 模拟请求参数
	// toUpbitParam := url.Values{}
	// toUpbitParam.Add("apiKey", apiKey)
	// toUpbitParam.Add("timestamp", fmt.Sprintf("%d", time.Now().UnixMilli()))

	// // 1. 生成 query string
	// queryString := toUpbitParam.Encode()

	// // 2. 从 PEM 加载 Ed25519 私钥
	// privKey, err := loadEd25519PrivateKey("myprivatekey.pem")
	// if err != nil {
	// 	panic(err)
	// }

	// // 3. 签名
	// signature := ed25519.Sign(privKey, []byte(queryString))

	// // 4. base64 编码
	// sigBase64 := base64.StdEncoding.EncodeToString(signature)

	// // 5. 拼接 URL
	// url := fmt.Sprintf("%s&signature=%s", queryString, sigBase64)

	// param := make(map[string]any)
	// //统一逻辑
	// param[p_API_KEY] = apiKey
	// param[p_TIME_STAMP] = timeUtils.GetNowTimeUnixMilli()
	// signRaw := buildQueryBytePool(128, param, querySortedKeyFast) //从池子中获取128位签名数据
	// fmt.Println(string(*signRaw))
	// signRes := byteBufPool.AcquireBuffer(64) //从池子中获取64位
	// defer byteBufPool.ReleaseBuffer(signRaw) //释放签名数据
	// defer byteBufPool.ReleaseBuffer(signRes) //释放签名值
	// if err := myCrypto.HmacSha256Fast(secretByte, *signRaw, signRes); err != nil {
	// 	return nil, err
	// }
	// return buildWsReqFast(512, api.Id, method, param, loginSortedKeyFast, signRes), nil
	return nil, nil
}

// NewFutureLoginSdk   rest查询订单 (USER_DATA)
func NewFutureLoginSdk() *FutureLoginSdk {
	return &FutureLoginSdk{}
}

func GetFutureLoginSdk(id string) *FutureLoginSdk {
	return NewFutureLoginSdk().Id_(id)
}

func loadEd25519PrivateKey(path string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM")
	}

	// 解析 PKCS#8 格式的私钥
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	priv, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an Ed25519 private key")
	}
	return priv, nil
}
