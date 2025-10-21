package orderModel

import "github.com/hhh500/upbitBnServer/internal/quant/market/symbolInfo/symbolStatic"

type MyPlaceOrderBatchReq struct {
	StaticMeta *symbolStatic.StaticTrade
	Orders     []*MyPlaceOrderReq
}

func NewMyPlaceOrderBatchReq(size int) *MyPlaceOrderBatchReq {
	return &MyPlaceOrderBatchReq{
		Orders: make([]*MyPlaceOrderReq, 0, size),
	}
}

func (s *MyPlaceOrderBatchReq) AddPlaceOrder(req *MyPlaceOrderReq) {
	s.Orders = append(s.Orders, req)
}

func (s *MyPlaceOrderBatchReq) GetClientOrderIds() (clientOrderIds []string) {
	clientOrderIds = make([]string, 0, len(s.Orders))
	for _, v := range s.Orders {
		clientOrderIds = append(clientOrderIds, v.ClientOrderId)
	}
	return
}
