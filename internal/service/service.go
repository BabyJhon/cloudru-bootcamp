package service

import (
	"net/http"
	"net/url"
	"time"

	"github.com/BabyJhon/cloudru-bootcamp/configs"
)

// Интерфейс для балансировки нагрузки
type Balancer interface {
	Next() *url.URL
	GetBackends() []*url.URL
}

type ClientIdentifier interface {
	IdentifyClient(r *http.Request) string
	GetAPIKey(r *http.Request) string
	GetClientIP(r *http.Request) string
}

type RateLimiterService interface {
	IsAllowed(clientID string) bool
	Stop()
}

type Service struct {
	Balancer         Balancer
	ClientIdentifier ClientIdentifier
	RateLimiter      RateLimiterService
	ClientService    *ClientService
}

func NewService(backends []*url.URL) *Service {
	cfg := configs.Load()

	config := RateLimiterConfig{
		DefaultCapacity: cfg.RateLimiter.Default.Capacity,
		DefaultRate:     cfg.RateLimiter.Default.RefillRate,
		RefillInterval:  time.Second,
	}

	rateLimiter := NewRateLimiter(config)

	clientService := NewClientService(rateLimiter)

	clientIdentifier := NewClientIdentifierService(true)

	for _, client := range cfg.RateLimiter.SpecialClients {
		rateLimiter.UpdateClient(client.ID, client.Capacity, client.RefillRate)
	}

	rateLimiter.SetIPBasedConfig(cfg.RateLimiter.IPBased.Capacity, cfg.RateLimiter.IPBased.RefillRate)

	return &Service{
		Balancer:         NewRoundRobinBalancer(backends),
		RateLimiter:      rateLimiter,
		ClientService:    clientService,
		ClientIdentifier: clientIdentifier,
	}
}
