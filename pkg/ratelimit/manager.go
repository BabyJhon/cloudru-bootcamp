package ratelimit

import (
	"sync"
	"time"
)

// TokenBucketManager централизованно управляет пополнением всех TokenBucket
type TokenBucketManager struct {
	buckets    map[string]*TokenBucket
	ticker     *time.Ticker
	refillTime time.Duration
	mu         sync.RWMutex
	stopCh     chan struct{}
}

func NewTokenBucketManager(refillInterval time.Duration) *TokenBucketManager {
	manager := &TokenBucketManager{
		buckets:    make(map[string]*TokenBucket),
		refillTime: refillInterval,
		stopCh:     make(chan struct{}),
	}
	
	manager.startRefillTicker()
	
	return manager
}

// фоновое пополнение токенов
func (m *TokenBucketManager) startRefillTicker() {
	m.ticker = time.NewTicker(m.refillTime)
	
	go func() {
		for {
			select {
			case <-m.ticker.C:
				m.refillAllBuckets()
			case <-m.stopCh:
				m.ticker.Stop()
				return
			}
		}
	}()
}

func (m *TokenBucketManager) refillAllBuckets() {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	tokensPerTick := m.refillTime.Seconds()
	
	// пополняем каждый бакет в соответствии с его скоростью
	for _, bucket := range m.buckets {
		tokensToAdd := tokensPerTick * bucket.refillRate
		bucket.AddTokens(tokensToAdd)
	}
}

func (m *TokenBucketManager) AddBucket(clientID string, bucket *TokenBucket) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.buckets[clientID] = bucket
}

func (m *TokenBucketManager) RemoveBucket(clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	delete(m.buckets, clientID)
}

func (m *TokenBucketManager) GetBucket(clientID string) (*TokenBucket, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	bucket, exists := m.buckets[clientID]
	return bucket, exists
}

func (m *TokenBucketManager) Stop() {
	close(m.stopCh)
}