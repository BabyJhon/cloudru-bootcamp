package ratelimit

import (
	"sync"
)

type TokenBucket struct {
	tokens     float64    
	capacity   int        
	refillRate float64    
	mu         sync.Mutex 
}

func NewTokenBucket(capacity int, refillRate float64) *TokenBucket {
	return &TokenBucket{
		tokens:     float64(capacity),
		capacity:   capacity,
		refillRate: refillRate,
	}
}

func (b *TokenBucket) TakeToken() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true
	}
	return false
}


func (b *TokenBucket) AddTokens(amount float64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.tokens = min(float64(b.capacity), b.tokens+amount)
}

func (b *TokenBucket) GetTokens() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.tokens
}

func (b *TokenBucket) UpdateRate(newRate float64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refillRate = newRate
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
