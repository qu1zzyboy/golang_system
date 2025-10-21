package coinMesh

import (
	"sort"
)

func (m *Manager) getSortDownCoins() []*CoinMesh {
	subs := make([]*CoinMesh, 0)
	m.coinMap.Range(func(key uint32, v *CoinMesh) bool {
		subs = append(subs, v)
		return true
	})
	sort.Sort(coinSortGtSlice(subs)) //订单价从大到小排序
	return subs
}

func (m *Manager) getSortRiseCoins() []*CoinMesh {
	subs := make([]*CoinMesh, 0)
	m.coinMap.Range(func(key uint32, v *CoinMesh) bool {
		subs = append(subs, v)
		return true
	})
	sort.Sort(coinSortLtSlice(subs)) //订单价从小到大排序
	return subs
}

func (m *Manager) GetSortCoins(isRise bool) []*CoinMesh {
	if isRise {
		return m.getSortRiseCoins()
	} else {
		return m.getSortDownCoins()
	}
}

// coinSortLtSlice 子订单排序从小到大
type coinSortLtSlice []*CoinMesh

func (tp coinSortLtSlice) Len() int {
	return len(tp)
}

func (tp coinSortLtSlice) Swap(i, j int) {
	tp[i], tp[j] = tp[j], tp[i]
}

func (tp coinSortLtSlice) Less(i, j int) bool {
	return tp[i].CmcRank < tp[j].CmcRank
}

// 子订单排序从大到小
type coinSortGtSlice []*CoinMesh

func (tp coinSortGtSlice) Len() int {
	return len(tp)
}

func (tp coinSortGtSlice) Swap(i, j int) {
	tp[i], tp[j] = tp[j], tp[i]
}

func (tp coinSortGtSlice) Less(i, j int) bool {
	return tp[i].CmcRank > tp[j].CmcRank
}
