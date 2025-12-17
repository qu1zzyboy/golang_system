package wsRequestCache

import (
	"time"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/instance/instanceDefine"
	"upbitBnServer/pkg/container/map/myMap"
	"upbitBnServer/pkg/singleton"
	"upbitBnServer/server/usageEnum"
)

type WsRequestType uint8

const (
	PLACE_ORDER WsRequestType = iota
	CANCEL_ORDER
	MODIFY_ORDER
	QUERY_ORDER
	QUERY_ACCOUNT_BALANCE
)

func (s WsRequestType) String() string {
	switch s {
	case PLACE_ORDER:
		return "PLACE_ORDER"
	case CANCEL_ORDER:
		return "CANCEL_ORDER"
	case MODIFY_ORDER:
		return "MODIFY_ORDER"
	case QUERY_ORDER:
		return "QUERY_ORDER"
	case QUERY_ACCOUNT_BALANCE:
		return "QUERY_ACCOUNT_BALANCE"
	}
	return "ERROR"
}

type WsRequestMeta struct {
	ReqJson       []byte
	ClientOrderId systemx.WsId16B     //订单的clientOrderId,只有订单数据才有
	UpdateAt      int64               //最后更新时间戳,毫秒
	ReqType       WsRequestType       //请求接口类型
	ReqFrom       instanceDefine.Type //实例枚举
	UsageFrom     usageEnum.Type      //用途枚举
}

type Manager struct {
	doJson *myMap.MySyncMap[systemx.WsId16B, *WsRequestMeta] //req_id-->订单json
}

var (
	cacheSingleton = singleton.NewSingleton(func() *Manager {
		return &Manager{
			doJson: myMap.NewMySyncMap[systemx.WsId16B, *WsRequestMeta](),
		}
	})
)

func GetCache() *Manager {
	return cacheSingleton.Get()
}

func (s *Manager) StoreMeta(reqId systemx.WsId16B, meta *WsRequestMeta) {
	s.doJson.Store(reqId, meta)
}

func (s *Manager) GetMeta(reqId systemx.WsId16B) (*WsRequestMeta, bool) {
	return s.doJson.Load(reqId)
}

func (s *Manager) DelMeta(reqId systemx.WsId16B) {
	s.doJson.Delete(reqId)
}

func init() {
	safex.SafeGo("start_check", func() {
		ticker_10s := time.NewTicker(10 * time.Second)
		const limit = 10000
		for range ticker_10s.C {
			ts := time.Now().UnixMilli()
			GetCache().doJson.Range(func(reqId systemx.WsId16B, meta *WsRequestMeta) bool {
				if ts-meta.UpdateAt > limit {
					dynamicLog.Error.GetLog().Errorf("[%s,%s]未接受返回,请求数据:%s", meta.ReqType, reqId, string(meta.ReqJson))
					GetCache().DelMeta(reqId)
				}
				return true
			})
		}
	})
}
