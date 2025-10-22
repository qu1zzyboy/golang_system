package sortStruct

import (
	"sort"

	"github.com/hhh500/quantGoInfra/pkg/container/map/myMap"
)

type Element struct {
	ClientOrderId  string
	OrderPriceSort float64 //订单价格,排序需要
}

type OneSideManager struct {
	subMap myMap.MySyncMap[string, *Element] //clientOrderId → 子订单
}

func NewOneSideManager() *OneSideManager {
	return &OneSideManager{
		subMap: myMap.NewMySyncMap[string, *Element](),
	}
}

func (s *OneSideManager) Store(clientOrderId string, orderPrice float64) {
	s.subMap.Store(clientOrderId, &Element{
		ClientOrderId:  clientOrderId,
		OrderPriceSort: orderPrice,
	})
}

func (s *OneSideManager) Delete(clientOrderId string) {
	s.subMap.Delete(clientOrderId)
}

func (s *OneSideManager) getSortBuyOrders() []*Element {
	subs := make([]*Element, 0)
	s.subMap.Range(func(key string, v *Element) bool {
		subs = append(subs, v)
		return true
	})
	sort.Sort(orderSortGtSlice(subs)) //订单价从大到小排序
	return subs
}

func (s *OneSideManager) getSortSellOrders() []*Element {
	subs := make([]*Element, 0)
	s.subMap.Range(func(key string, v *Element) bool {
		subs = append(subs, v)
		return true
	})
	sort.Sort(orderSortLtSlice(subs)) //订单价从小到大排序
	return subs
}

func (s *OneSideManager) GetSortOrders(isBuy bool) []*Element {
	if isBuy {
		return s.getSortBuyOrders()
	} else {
		return s.getSortSellOrders()
	}
}

func (s *OneSideManager) GetLastSortOrder(isBuy bool) *Element {
	subs := s.GetSortOrders(isBuy)
	if len(subs) == 0 {
		return nil
	}
	return (subs)[len(subs)-1]
}

// orderSortLtSlice 子订单排序从小到大
type orderSortLtSlice []*Element

func (tp orderSortLtSlice) Len() int {
	return len(tp)
}

func (tp orderSortLtSlice) Swap(i, j int) {
	tp[i], tp[j] = tp[j], tp[i]
}

func (tp orderSortLtSlice) Less(i, j int) bool {
	return tp[i].OrderPriceSort < tp[j].OrderPriceSort
}

// 子订单排序从大到小
type orderSortGtSlice []*Element

func (tp orderSortGtSlice) Len() int {
	return len(tp)
}

func (tp orderSortGtSlice) Swap(i, j int) {
	tp[i], tp[j] = tp[j], tp[i]
}

func (tp orderSortGtSlice) Less(i, j int) bool {
	return tp[i].OrderPriceSort > tp[j].OrderPriceSort
}
