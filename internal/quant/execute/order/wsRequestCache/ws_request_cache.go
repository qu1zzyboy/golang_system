package wsRequestCache

import (
	"upbitBnServer/internal/quant/execute/order/orderBelongEnum"
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
	Json    string
	ReqType WsRequestType
	ReqFrom orderBelongEnum.Type
}

type Manager struct {
	doJson myMap.MySyncMap[string, *WsRequestMeta] //req_id-->订单json
}

var (
	cacheSingleton = singleton.NewSingleton(func() *Manager {
		return &Manager{
			doJson: myMap.NewMySyncMap[string, *WsRequestMeta](),
		}
	})
)

func GetCache() *Manager {
	return cacheSingleton.Get()
}

func (s *Manager) StoreMeta(clientOrderId string, meta *WsRequestMeta) {
	s.doJson.Store(clientOrderId, meta)
}

func (s *Manager) GetMeta(clientOrderId string) (*WsRequestMeta, bool) {
	return s.doJson.Load(clientOrderId)
}

func (s *Manager) DelMeta(clientOrderId string) {
	s.doJson.Delete(clientOrderId)
}
func (s *Manager) Len() int {
	return s.doJson.Length()
}
