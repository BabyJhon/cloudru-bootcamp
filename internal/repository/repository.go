package repository

import (
	"context"

	"github.com/BabyJhon/cloudru-bootcamp/internal/entity"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client interface {
	CreateClient(ctx context.Context, client *entity.RateLimitClient) (int, error)
	GetClient(ctx context.Context, clientID string) (*entity.RateLimitClient, error)
	UpdateClient(ctx context.Context, client *entity.RateLimitClient) (int, error)
	DeleteClient(ctx context.Context, clientID string) (int, error)
	ListClients(ctx context.Context) ([]*entity.RateLimitClient, error)
}

type Repository struct {
	Client
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		Client: NewClientRepo(db),
	}
}
