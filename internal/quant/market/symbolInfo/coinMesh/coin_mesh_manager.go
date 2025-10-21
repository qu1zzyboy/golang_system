package coinMesh

import (
	"fmt"
	"os"
	"sort"

	"github.com/hhh500/quantGoInfra/pkg/container/map/myMap"
	"github.com/hhh500/quantGoInfra/pkg/singleton"
	"github.com/hhh500/quantGoInfra/pkg/utils/convertx"
)

const REDIS_KEY_COIN_MESH_ALL = "COIN_MESH_ALL"

var serviceSingleton = singleton.NewSingleton(func() *Manager { return &Manager{coinMap: myMap.NewMySyncMap[uint32, *CoinMesh]()} })

func GetManager() *Manager { return serviceSingleton.Get() }

type Manager struct {
	coinMap myMap.MySyncMap[uint32, *CoinMesh] //cmcId-->StaticTrade
}

func (m *Manager) Set(mesh *CoinMesh) {
	m.coinMap.Store(mesh.CmcId, mesh)
}

func (m *Manager) Get(cmcId uint32) (*CoinMesh, bool) {
	return m.coinMap.Load(cmcId)
}

func (m *Manager) GetLength() int {
	return m.coinMap.Length()
}

func (m *Manager) PrintAll() {
	// Step 1: 收集所有 CoinMesh 到切片中
	var list []*CoinMesh
	m.coinMap.Range(func(_ uint32, value *CoinMesh) bool {
		list = append(list, value)
		return true
	})
	// Step 2: 按 CmcAsset 排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].CmcRank < list[j].CmcRank
	})
	for _, va := range list {
		fmt.Printf("%d CmcAsset=%-10s |cap=%.2f | BnFutureUsdt=%-15s  | BnSpotUsdt=%-12s | BybitFutureUsdt=%-14s | UpbitSpotKrw=%-12s  [%s,%d] CmcDesc=%-12s\n",
			va.CmcRank,
			va.CmcAsset,
			va.MarketCap,
			va.BnFuUsdtName,
			va.BnSpUsdtName,
			va.BybitFuUsdtName,
			va.UpbitSpKrwName,
			va.CmcDateAddedStr,
			va.CmcDateAdded,
			va.CmcDesc,
		)
	}
	fmt.Println("Total CoinMesh Count:", m.coinMap.Length())
	fmt.Println()
}

func (m *Manager) WriteToCsv(csvPath string) {
	// Step 1: 收集所有 CoinMesh 到切片中
	var list []*CoinMesh
	m.coinMap.Range(func(_ uint32, value *CoinMesh) bool {
		list = append(list, value)
		return true
	})
	// Step 2: 按 CmcAsset 排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].CmcRank < list[j].CmcRank
	})
	var records [][]string
	records = append(records, []string{
		"cmcRank",
		"CmcAsset",
		"marketCap",
		"bn Usdt合约",
		"bn Usdc合约",
		"bn Usdt现货",
		"bybit Usdt合约",
		"upbit Krw现货",
		"cmc添加时间",
		"cmc添加时间戳",
		"描述",
	})
	for _, va := range list {
		records = append(records, []string{
			convertx.ToString(va.CmcRank),
			va.CmcAsset,
			convertx.ToString(va.MarketCap),
			va.BnFuUsdtName,
			va.BnFuUsdcName,
			va.BnSpUsdtName,
			va.BybitFuUsdtName,
			va.UpbitSpKrwName,
			va.CmcDateAddedStr,
			convertx.ToString(va.CmcDateAdded),
			va.CmcDesc,
		})
	}
	os.Remove(csvPath) // 删除旧文件
	// fileUtils.AppendDataToDesCsv(records, csvPath)
}

func (m *Manager) DelEmpty() {
	m.coinMap.Range(func(cmcId uint32, value *CoinMesh) bool {
		if value.BnFuUsdtName == "" && value.BnFuUsdcName == "" && value.BnSpUsdtName == "" &&
			value.BybitFuUsdtName == "" &&
			value.UpbitSpKrwName == "" {
			m.coinMap.Delete(cmcId)
		}
		return true
	})
}

func (m *Manager) DeepCopy() (copy *Manager, copyMap map[uint32]*CoinMesh) {
	copy = &Manager{coinMap: myMap.NewMySyncMap[uint32, *CoinMesh]()}
	copyMap = make(map[uint32]*CoinMesh, m.coinMap.Length())
	m.coinMap.Range(func(cmcId uint32, value *CoinMesh) bool {
		meshCopy := *value
		copy.coinMap.Store(cmcId, &meshCopy)
		copyMap[cmcId] = &meshCopy
		return true
	})
	return copy, copyMap
}

func (m *Manager) Del(cmcId uint32) {
	m.coinMap.Delete(cmcId)
}
