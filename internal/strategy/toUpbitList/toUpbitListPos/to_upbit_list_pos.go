package toUpbitListPos

import (
	"maps"
	"sync"

	"github.com/shopspring/decimal"
)

type PosCal struct {
	posRw          sync.RWMutex              // 仓位读写锁
	posMap         map[uint8]decimal.Decimal // 各个账户仓位情况
	posTotalAmount decimal.Decimal           // 累计仓位情况
}

func NewPosCal() *PosCal {
	return &PosCal{
		posMap: make(map[uint8]decimal.Decimal),
	}
}

func (s *PosCal) OpenFilled(accountKeyId uint8, volume decimal.Decimal) decimal.Decimal {
	s.posRw.Lock()
	defer s.posRw.Unlock()
	s.posTotalAmount = s.posTotalAmount.Add(volume)
	if pos, ok := s.posMap[accountKeyId]; ok {
		s.posMap[accountKeyId] = pos.Add(volume)
	} else {
		s.posMap[accountKeyId] = volume
	}
	return s.posTotalAmount
}

func (s *PosCal) CloseFilled(accountKeyId uint8, volume decimal.Decimal) decimal.Decimal {
	s.posRw.Lock()
	defer s.posRw.Unlock()
	s.posTotalAmount = s.posTotalAmount.Sub(volume)
	if pos, ok := s.posMap[accountKeyId]; ok {
		s.posMap[accountKeyId] = pos.Sub(volume)
	}
	return s.posTotalAmount
}

func (s *PosCal) GetTotal() decimal.Decimal {
	s.posRw.RLock()
	defer s.posRw.RUnlock()
	return s.posTotalAmount
}

func (s *PosCal) Clear() {
	s.posRw.Lock()
	defer s.posRw.Unlock()
	s.posTotalAmount = decimal.Zero
	s.posMap = make(map[uint8]decimal.Decimal)
}

func (s *PosCal) GetAllAccountPos() map[uint8]decimal.Decimal {
	s.posRw.RLock()
	defer s.posRw.RUnlock()
	copyMap := make(map[uint8]decimal.Decimal, len(s.posMap))
	maps.Copy(copyMap, s.posMap)
	return copyMap
}
