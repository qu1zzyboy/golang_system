package orderStatic

import (
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"upbitBnServer/pkg/container/map/myMap"
	"upbitBnServer/pkg/singleton"

	"github.com/shopspring/decimal"
)

//订单解析流程,静态数据能复用就复用,不能复用就重新解析
//rest知道数据是属于哪个接口的
//改单接口是例外的,ws数据先判断有没有被改过

type StaticMeta struct {
	OrigPrice   decimal.Decimal      // 原始价格
	OrigVolume  decimal.Decimal      // 原始数量
	SymbolIndex int                  // 交易对的唯一标识
	IsModified  bool                 // 标记是否被修改过,被修改过要重新解析原始价格和数量
	OrderMode   execute.MyOrderMode  // 订单模式
	OrderFrom   orderBelongEnum.Type // 实例id枚举
	_           [16]byte             // 填充到 64 字节
}

type Service struct {
	metaMap myMap.MySyncMap[string, StaticMeta] // clientOrderId --> StaticMeta
}

var serviceSingleton = singleton.NewSingleton(func() *Service {
	return &Service{metaMap: myMap.NewMySyncMap[string, StaticMeta]()}
})

func GetService() *Service {
	return serviceSingleton.Get()
}

func (s *Service) SaveOrderMeta(clientOrderId string, meta StaticMeta) {
	s.metaMap.Store(clientOrderId, meta)
}

func (s *Service) RemoveOrderMeta(clientOrderId string) {
	s.metaMap.Delete(clientOrderId)
}

func (s *Service) GetOrderMeta(clientOrderId string) (StaticMeta, bool) {
	if meta, ok := s.metaMap.Load(clientOrderId); ok {
		return meta, true
	}
	return StaticMeta{}, false
}

func (s *Service) IsOrderExist(clientOrderId string) (isExist bool, isModified bool) {
	if meta, ok := s.metaMap.Load(clientOrderId); ok {
		return true, meta.IsModified
	}
	return false, false
}

func (s *Service) GetIsModified(clientOrderId string) bool {
	if meta, ok := s.GetOrderMeta(clientOrderId); ok {
		return meta.IsModified
	}
	return false
}

func (s *Service) GetOrderInstanceId(clientOrderId string) (orderBelongEnum.Type, execute.MyOrderMode, bool) {
	if meta, ok := s.GetOrderMeta(clientOrderId); ok {
		return meta.OrderFrom, meta.OrderMode, true
	}
	return orderBelongEnum.UNKNOWN, execute.ORDER_MODE_ERROR, false
}

func (s *Service) GetOrderInstanceIdAndSymbolId(clientOrderId string) (orderBelongEnum.Type, execute.MyOrderMode, int, bool) {
	if meta, ok := s.GetOrderMeta(clientOrderId); ok {
		return meta.OrderFrom, meta.OrderMode, meta.SymbolIndex, true
	}
	return orderBelongEnum.UNKNOWN, execute.ORDER_MODE_ERROR, 0, false
}
