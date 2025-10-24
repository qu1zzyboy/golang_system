package myMap

import (
	"math"
	"sort"
	"sync"

	"github.com/shopspring/decimal"
)

type SyncFloat64Map struct {
	syncMap MySyncMap[string, float64]
}

func NewSyncFloat64Map() *SyncFloat64Map {
	return &SyncFloat64Map{
		syncMap: NewMySyncMap[string, float64](),
	}
}

func (s *SyncFloat64Map) Store(key string, value float64) {
	s.syncMap.Store(key, value)
}

func (s *SyncFloat64Map) ClearMap() {
	s.syncMap.Range(func(key string, v float64) bool {
		s.syncMap.Delete(key)
		return true
	})
}

func (s *SyncFloat64Map) Load(key string) (value float64, ok bool) {
	value, ok = s.syncMap.Load(key)
	if !ok {
		return math.NaN(), false
	}
	return value, true
}

func (s *SyncFloat64Map) Length() int {
	length := 0
	s.syncMap.Range(func(k string, v float64) bool {
		length += 1
		return true
	})
	return length
}

func (s *SyncFloat64Map) GetMap() map[string]float64 {
	resp := make(map[string]float64)
	s.syncMap.Range(func(key string, v float64) bool {
		resp[key] = v
		return true
	})
	return resp
}

type kv struct {
	Key   string
	Value float64
}

// Rank 按value降序排序key
func (s *SyncFloat64Map) Rank() []string {
	var ss []kv
	s.syncMap.Range(func(key string, value float64) bool {
		ss = append(ss, kv{key, value})
		return true
	})
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value // 按value升序排序
	})
	var Slice []string
	for _, tmp := range ss {
		Slice = append(Slice, tmp.Key)
	}
	return Slice
}

/*====================================================================*/

type Counter struct {
	value decimal.Decimal
	mu    sync.Mutex
}

func NewCounter() *Counter { return &Counter{value: decimal.NewFromFloat(0.0)} }

func (c *Counter) Increment(incr decimal.Decimal) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value = c.value.Add(incr)
}

func (c *Counter) Value() decimal.Decimal {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

func (c *Counter) Decrement(incr decimal.Decimal) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value = c.value.Sub(incr)
}
