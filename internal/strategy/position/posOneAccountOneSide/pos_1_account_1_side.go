package posOneAccountOneSide

import (
	"sync"
)

type PosCalSafe struct {
	posRw          sync.RWMutex // 仓位读写锁
	posTotalAmount float64      // 累计仓位情况
}

func NewPosCal() *PosCalSafe {
	return &PosCalSafe{}
}

func (s *PosCalSafe) OpenFilled(accountKeyId uint8, volume float64) float64 {
	s.posRw.Lock()
	defer s.posRw.Unlock()
	s.posTotalAmount = s.posTotalAmount + volume
	return s.posTotalAmount
}

func (s *PosCalSafe) CloseFilled(accountKeyId uint8, volume float64) float64 {
	s.posRw.Lock()
	defer s.posRw.Unlock()
	s.posTotalAmount = s.posTotalAmount - volume
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
}
