package wsRequestCache

import (
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/instanceEnum"
	"upbitBnServer/internal/infra/systemx/usageEnum"
	"upbitBnServer/pkg/container/map/myMap"
	"upbitBnServer/pkg/singleton"
)

type WsRequestType uint8

const (
	PLACE_ORDER WsRequestType = iota
	CANCEL_ORDER
	MODIFY_ORDER
	QUERY_ORDER
	QUERY_ACCOUNT_BALANCE
)

type WsRequestMeta struct {
	ReqJson   []byte
	ReqType   WsRequestType     //请求接口类型
	ReqFrom   instanceEnum.Type //实例枚举
	UsageFrom usageEnum.Type    //用途枚举
}

type Manager struct {
	doJson myMap.MySyncMap[systemx.WsId16B, WsRequestMeta] //req_id-->订单json
}

var (
	cacheSingleton = singleton.NewSingleton(func() *Manager {
		return &Manager{
			doJson: myMap.NewMySyncMap[systemx.WsId16B, WsRequestMeta](),
		}
	})
)

func GetCache() *Manager {
	return cacheSingleton.Get()
}

func (s *Manager) StoreMeta(reqId systemx.WsId16B, meta WsRequestMeta) {
	s.doJson.Store(reqId, meta)
}

func (s *Manager) GetMeta(reqId systemx.WsId16B) (WsRequestMeta, bool) {
	return s.doJson.Load(reqId)
}

func (s *Manager) DelMeta(reqId systemx.WsId16B) {
	s.doJson.Delete(reqId)
}
