package posMultiAccountOneSide

import (
	"maps"
	"sync"
)

type PosCalSafe struct {
	posRw          sync.RWMutex      // 仓位读写锁
	posMap         map[uint8]float64 // 各个账户仓位情况
	posTotalAmount float64           // 累计仓位情况
}

func NewPosCal() *PosCalSafe {
	return &PosCalSafe{
		posMap: make(map[uint8]float64),
	}
}

func (s *PosCalSafe) OpenFilled(accountKeyId uint8, volume float64) float64 {
	s.posRw.Lock()
	defer s.posRw.Unlock()
	s.posTotalAmount = s.posTotalAmount + volume
	if pos, ok := s.posMap[accountKeyId]; ok {
		s.posMap[accountKeyId] = pos + volume
	} else {
		s.posMap[accountKeyId] = volume
	}
	return s.posTotalAmount
}

func (s *PosCalSafe) CloseFilled(accountKeyId uint8, volume float64) float64 {
	s.posRw.Lock()
	defer s.posRw.Unlock()
	s.posTotalAmount = s.posTotalAmount - volume
	if pos, ok := s.posMap[accountKeyId]; ok {
		s.posMap[accountKeyId] = pos - volume
	}
	return s.posTotalAmount
}

func (s *PosCalSafe) GetTotal() float64 {
	s.posRw.RLock()
	defer s.posRw.RUnlock()
	return s.posTotalAmount
}

func (s *PosCalSafe) Clear() {
	s.posRw.Lock()
	defer s.posRw.Unlock()
	s.posTotalAmount = 0
	s.posMap = make(map[uint8]float64)
}

func (s *PosCalSafe) GetAllAccountPos() map[uint8]float64 {
	s.posRw.RLock()
	defer s.posRw.RUnlock()
	copyMap := make(map[uint8]float64, len(s.posMap))
	maps.Copy(copyMap, s.posMap)
	return copyMap
}
