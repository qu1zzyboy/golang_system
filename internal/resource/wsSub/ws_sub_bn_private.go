package wsSub

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"upbitBnServer/internal/define/defineEmoji"
	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/internal/quant/execute/order/orderSdk/bn/orderSdkBnRest"

	"github.com/gorilla/websocket"
)

const listenKeyRefreshInterval = 20 * time.Minute //  listenKey刷新间隔

type BnPayload struct {
	baseUrl   string                     // 当前订阅的URL
	listenKey string                     // 监听密钥
	once      sync.Once                  // 确保只创建一次连接
	restFu    *orderSdkBnRest.FutureRest //rest client
	conn      *wsDefine.SafeWrite        // websocket连接
}

func NewBnPayload(apiKey, secretKey string) *BnPayload {
	return &BnPayload{
		baseUrl: "wss://fstream.binance.com/ws/",
		restFu:  orderSdkBnRest.NewFutureRest(apiKey, secretKey),
	}
}

func (s *BnPayload) createListenKey() error {
	res, err := s.restFu.DoListenKey()
	if err != nil || res == "" {
		return err
	}
	s.listenKey = res
	return nil
}

func (s *BnPayload) refreshListenKey(url string) {
	s.once.Do(func() {
		dynamicLog.Log.GetLog().Infof("进入[%s] refresh循环 %s", url, defineEmoji.Rocket)
		safex.SafeGo(url, func() {
			ticker := time.NewTicker(listenKeyRefreshInterval)
			defer ticker.Stop()
			for range ticker.C {
				err := s.restFu.DelayListenKey(s.listenKey)
				for err != nil {
					time.Sleep(5 * time.Second)
					if strings.Contains(err.Error(), "-1125") {
						err = s.createListenKey() //如果是-1125错误,则Post更新
					} else {
						err = s.restFu.DelayListenKey(s.listenKey)
					}
				}
			}
		})
	})
}

func (s *BnPayload) DialTo(ctx context.Context) (*wsDefine.SafeWrite, error) {
	if err := s.createListenKey(); err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s%s", s.baseUrl, s.listenKey)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, connErr.WithCause(err).WithMetadata(map[string]string{defineJson.FullUrl: url})
	}
	s.conn = wsDefine.NewSafeWrite(conn)
	s.refreshListenKey(url) // 启动监听密钥刷新协程
	return s.conn, nil
}

func (s *BnPayload) GetUrl() string {
	return s.baseUrl
}
