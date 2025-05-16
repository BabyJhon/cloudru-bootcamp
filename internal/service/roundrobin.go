package service

import (
	"net/url"
	"sync"
	"sync/atomic"
)

type RoundRobinBalancer struct {
	backends []*url.URL
	current  uint32
	mutex    sync.RWMutex
}

func NewRoundRobinBalancer(backends []*url.URL) *RoundRobinBalancer {
	return &RoundRobinBalancer{
		backends: backends,
		current:  0,
	}
}

// добавляет бэкенд в балансировщик потокобезопасным способом
func (b *RoundRobinBalancer) AddBackend(backend *url.URL) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.backends = append(b.backends, backend)
}

func (b *RoundRobinBalancer) RemoveBackend(backend *url.URL) bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for i, u := range b.backends {
		if u.String() == backend.String() {
			b.backends = append(b.backends[:i], b.backends[i+1:]...)
			return true
		}
	}
	return false
}

func (b *RoundRobinBalancer) Next() *url.URL {
	next := atomic.AddUint32(&b.current, 1)

	// захватываем блокировку на чтение для доступа к списку бэкендов
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	if len(b.backends) == 0 {
		return nil
	}

	return b.backends[(next-1)%uint32(len(b.backends))]
}

func (b *RoundRobinBalancer) GetBackends() []*url.URL {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	// делаем копию слайса для предотвращения гонок данных
	result := make([]*url.URL, len(b.backends))
	copy(result, b.backends)

	return result
}
