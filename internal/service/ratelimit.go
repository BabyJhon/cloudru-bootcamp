package service

import (
	"strings"
	"sync"
	"time"

	"github.com/BabyJhon/cloudru-bootcamp/internal/entity"
	"github.com/BabyJhon/cloudru-bootcamp/pkg/ratelimit"
)

// RateLimiterConfig содержит настройки для ограничителя нагрузки
type RateLimiterConfig struct {
	DefaultCapacity int           
	DefaultRate     float64       // Скорость измеряется в токенах/сек
	RefillInterval  time.Duration 
	IPBasedCapacity int           // Настройки для IP-based ограничения
	IPBasedRate     float64       
}

type RateLimiter struct {
	bucketManager *ratelimit.TokenBucketManager
	config        RateLimiterConfig
	clients       *sync.Map // Хранилище настроек клиентов (IP/API-ключей)
	stopCh        chan struct{}
}

func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	return &RateLimiter{
		bucketManager: ratelimit.NewTokenBucketManager(config.RefillInterval),
		config:        config,
		clients:       &sync.Map{},
		stopCh:        make(chan struct{}),
	}
}

func (s *RateLimiter) Stop() {
	close(s.stopCh)
	s.bucketManager.Stop()
}

// Возвращает true, если запрос разрешен (есть токен) и false, если запрос следует отклонить
func (s *RateLimiter) IsAllowed(clientID string) bool {
	bucket, exists := s.bucketManager.GetBucket(clientID)
	if !exists {
		s.getOrCreateClient(clientID)
		bucket, _ = s.bucketManager.GetBucket(clientID)
	}

	return bucket.TakeToken()
}

func (s *RateLimiter) getOrCreateClient(clientID string) *entity.RateLimitClient {
	if clientValue, found := s.clients.Load(clientID); found {
		return clientValue.(*entity.RateLimitClient)
	}

	isIPBased := strings.HasPrefix(clientID, "ip:")

	client := &entity.RateLimitClient{
		ID: clientID,
	}

	if isIPBased {
		client.Capacity = s.config.IPBasedCapacity
		client.RefillRate = s.config.IPBasedRate
	} else {
		client.Capacity = s.config.DefaultCapacity
		client.RefillRate = s.config.DefaultRate
	}

	// cоздаем токен-бакет и добавляем его в менеджер
	bucket := ratelimit.NewTokenBucket(client.Capacity, client.RefillRate)
	s.bucketManager.AddBucket(clientID, bucket)

	// cохраняем в памяти
	s.clients.Store(clientID, client)

	return client
}

func (s *RateLimiter) GetClient(clientID string) (*entity.RateLimitClient, bool) {
	clientValue, found := s.clients.Load(clientID)
	if !found {
		return nil, false
	}
	return clientValue.(*entity.RateLimitClient), true
}

func (s *RateLimiter) UpdateClient(clientID string, capacity int, ratePerSec float64) {
	client, exists := s.GetClient(clientID)

	if exists {
		client.Capacity = capacity
		client.RefillRate = ratePerSec

		if bucket, found := s.bucketManager.GetBucket(clientID); found {
			bucket.UpdateRate(ratePerSec)
		}
	} else {
		client = &entity.RateLimitClient{
			ID:         clientID,
			Capacity:   capacity,
			RefillRate: ratePerSec,
		}
		bucket := ratelimit.NewTokenBucket(capacity, ratePerSec)
		s.bucketManager.AddBucket(clientID, bucket)
	}

	s.clients.Store(clientID, client)
}

func (s *RateLimiter) DeleteClient(clientID string) {
	// удаляем бакет из менеджера и настройки из памяти
	s.bucketManager.RemoveBucket(clientID)
	s.clients.Delete(clientID)
}

func (s *RateLimiter) ListClients() []entity.RateLimitClient {
	var clients []entity.RateLimitClient

	s.clients.Range(func(_, value interface{}) bool {
		client := value.(*entity.RateLimitClient)
		clients = append(clients, *client)
		return true
	})

	return clients
}

func (s *RateLimiter) GetTokensRemaining(clientID string) (float64, bool) {
	bucket, exists := s.bucketManager.GetBucket(clientID)
	if !exists {
		return 0, false
	}
	return bucket.GetTokens(), true
}

func (s *RateLimiter) SetIPBasedConfig(capacity int, ratePerSec float64) {
	s.config.IPBasedCapacity = capacity
	s.config.IPBasedRate = ratePerSec
}
