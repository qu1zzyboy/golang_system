package orderStatic

import (
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/instanceEnum"
	"upbitBnServer/internal/infra/systemx/usageEnum"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/pkg/container/map/myMap"
	"upbitBnServer/pkg/singleton"
)

//订单解析流程,静态数据能复用就复用,不能复用就重新解析
//rest知道数据是属于哪个接口的
//改单接口是例外的,ws数据先判断有没有被改过

type StaticMeta struct {
	SymbolIndex systemx.SymbolIndex16I // 交易对的唯一标识
	Pvalue      uint64                 //定点价格
	Qvalue      uint64                 //定点数量
	SymbolLen   uint16                 // 交易对长度
	IsModified  bool                   // 标记是否被修改过,被修改过要重新解析原始价格和数量
	OrderMode   execute.OrderMode      // 订单模式
	ReqFrom     instanceEnum.Type      //实例枚举
	UsageFrom   usageEnum.Type         //用途枚举
}

type Service struct {
	metaMap myMap.MySyncMap[systemx.WsId16B, StaticMeta] // clientOrderId --> StaticMeta
}

var serviceSingleton = singleton.NewSingleton(func() *Service {
	return &Service{metaMap: myMap.NewMySyncMap[systemx.WsId16B, StaticMeta]()}
})

func GetService() *Service {
	return serviceSingleton.Get()
}

func (s *Service) SaveOrderMeta(clientOrderId systemx.WsId16B, meta StaticMeta) {
	s.metaMap.Store(clientOrderId, meta)
}

func (s *Service) DelOrderMeta(clientOrderId systemx.WsId16B) {
	s.metaMap.Delete(clientOrderId)
}

func (s *Service) GetOrderMeta(clientOrderId systemx.WsId16B) (StaticMeta, bool) {
	if meta, ok := s.metaMap.Load(clientOrderId); ok {
		return meta, true
	}
	return StaticMeta{}, false
}

func (s *Service) IsOrderExist(clientOrderId systemx.WsId16B) (isExist bool, isModified bool) {
	if meta, ok := s.metaMap.Load(clientOrderId); ok {
		return true, meta.IsModified
	}
	return false, false
}

func (s *Service) GetIsModified(clientOrderId systemx.WsId16B) bool {
	if meta, ok := s.GetOrderMeta(clientOrderId); ok {
		return meta.IsModified
	}
	return false
}
