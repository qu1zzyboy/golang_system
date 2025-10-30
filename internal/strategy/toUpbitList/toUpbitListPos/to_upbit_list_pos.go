package toUpbitListPos

import (
	"maps"
	"sync"

	"github.com/shopspring/decimal"
)

type PosCal struct {
	posRw       sync.RWMutex              // 仓位读写锁
	posMap      map[uint8]decimal.Decimal // 各个账户仓位情况
	totalAmount decimal.Decimal           // 累计 num 情况
	totalQty    decimal.Decimal           // 累计 qty 情况
	totalAvg    decimal.Decimal           // 累计 avg 情况
}

func NewPosCal() *PosCal {
	return &PosCal{
		posMap: make(map[uint8]decimal.Decimal),
	}
}

func (s *PosCal) OpenFilled(accountKeyId uint8, avg, volume decimal.Decimal) (decimal.Decimal, float64) {
	s.posRw.Lock()
	defer s.posRw.Unlock()
	s.totalAmount = s.totalAmount.Add(volume)
	s.totalQty = s.totalQty.Add(avg.Mul(volume)) //qty=avg*vol
	s.totalAvg = s.totalQty.Div(s.totalAmount)   //avg=qty/vol
	if pos, ok := s.posMap[accountKeyId]; ok {
		s.posMap[accountKeyId] = pos.Add(volume)
	} else {
		s.posMap[accountKeyId] = volume
	}
	return s.totalAmount, s.totalAvg.InexactFloat64()
}

func (s *PosCal) CloseFilled(accountKeyId uint8, volume decimal.Decimal) decimal.Decimal {
	s.posRw.Lock()
	defer s.posRw.Unlock()
	s.totalAmount = s.totalAmount.Sub(volume)
	if pos, ok := s.posMap[accountKeyId]; ok {
		s.posMap[accountKeyId] = pos.Sub(volume)
	}
	return s.totalAmount
}

func (s *PosCal) GetTotalVol() decimal.Decimal {
	s.posRw.RLock()
	defer s.posRw.RUnlock()
	return s.totalAmount
}

func (s *PosCal) Clear() {
	s.posRw.Lock()
	defer s.posRw.Unlock()
	s.totalAmount = decimal.Zero
	s.posMap = make(map[uint8]decimal.Decimal)
}

func (s *PosCal) GetAllAccountPos() map[uint8]decimal.Decimal {
	s.posRw.RLock()
	defer s.posRw.RUnlock()
	copyMap := make(map[uint8]decimal.Decimal, len(s.posMap))
	maps.Copy(copyMap, s.posMap)
	return copyMap
}
