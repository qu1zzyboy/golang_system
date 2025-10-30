package orderModel

import "upbitBnServer/internal/quant/market/symbolInfo/symbolStatic"

type MyCancelOrderBatchReq struct {
	ApiFrom    string //打印字段,api来源
	size       int
	Orders     []*MyQueryOrderReq
	StaticMeta *symbolStatic.StaticTrade
}

func NewMyCancelOrderBatchReq(size int) *MyCancelOrderBatchReq {
	return &MyCancelOrderBatchReq{
		size:   size,
		Orders: make([]*MyQueryOrderReq, 0, size),
	}
}

func (s *MyCancelOrderBatchReq) AddCancelOrderArray(req []*MyQueryOrderReq) {
	s.Orders = append(s.Orders, req...)
}

func (s *MyCancelOrderBatchReq) AddCancelOrder(req *MyQueryOrderReq) {
	s.Orders = append(s.Orders, req)
}

func (s *MyCancelOrderBatchReq) GetClientOrderIds() (clientOrderIds []string) {
	clientOrderIds = make([]string, 0, s.size)
	for _, v := range s.Orders {
		clientOrderIds = append(clientOrderIds, v.ClientOrderId)
	}
	return
}
