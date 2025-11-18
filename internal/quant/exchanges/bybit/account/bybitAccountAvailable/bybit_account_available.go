package bybitAccountAvailable

import (
	"sync"
	"upbitBnServer/pkg/singleton"
)

var serviceSingleton = singleton.NewSingleton(func() *Manager {
	return &Manager{}
})

func GetManager() *Manager {
	return serviceSingleton.Get()
}

type Manager struct {
	posRw    sync.RWMutex // 仓位读写锁
	accounts []float64
}

func (m *Manager) init(length int) {
	m.accounts = make([]float64, length)
}

func (m *Manager) SetAvailable(accountKeyId uint8, available float64) {
	m.posRw.Lock()
	defer m.posRw.Unlock()
	m.accounts[accountKeyId] = available
}

func (m *Manager) GetAvailable(accountKeyId uint8) float64 {
	m.posRw.RLock()
	defer m.posRw.RUnlock()
	return m.accounts[accountKeyId]
}
