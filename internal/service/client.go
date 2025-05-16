package service

import (
	"errors"

	"github.com/BabyJhon/cloudru-bootcamp/internal/entity"
)

type ClientService struct {
	rateLimiter *RateLimiter
}

func NewClientService(rateLimiter *RateLimiter) *ClientService {
	return &ClientService{
		rateLimiter: rateLimiter,
	}
}

func (s *ClientService) CreateClient(req *entity.CreateClientRequest) error {
	_, exists := s.rateLimiter.GetClient(req.ClientID)
	if exists {
		return errors.New("client already exists")
	}

	s.rateLimiter.UpdateClient(req.ClientID, req.Capacity, req.RatePerSec)
	return nil
}

func (s *ClientService) GetClient(clientID string) (*entity.RateLimitClient, error) {
	client, exists := s.rateLimiter.GetClient(clientID)
	if !exists {
		return nil, errors.New("client not found")
	}
	return client, nil
}

func (s *ClientService) UpdateClient(clientID string, req *entity.UpdateClientRequest) error {
	_, exists := s.rateLimiter.GetClient(clientID)
	if !exists {
		return errors.New("client not found")
	}

	s.rateLimiter.UpdateClient(clientID, req.Capacity, req.RatePerSec)
	return nil
}

func (s *ClientService) DeleteClient(clientID string) error {
	_, exists := s.rateLimiter.GetClient(clientID)
	if !exists {
		return errors.New("client not found")
	}

	s.rateLimiter.DeleteClient(clientID)
	return nil
}

func (s *ClientService) ListClients() entity.ClientList {
	clients := s.rateLimiter.ListClients()
	return entity.ClientList{
		Clients: clients,
		Total:   len(clients),
	}
}

func (s *ClientService) GetTokensRemaining(clientID string) (float64, error) {
	tokens, exists := s.rateLimiter.GetTokensRemaining(clientID)
	if !exists {
		return 0, errors.New("client not found")
	}
	return tokens, nil
}
