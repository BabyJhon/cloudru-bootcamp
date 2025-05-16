package repository

import (
	"context"
	"errors"

	"github.com/BabyJhon/cloudru-bootcamp/internal/entity"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRateLimitRepository реализует хранилище в PostgreSQL
type ClientRepo struct {
	db *pgxpool.Pool
}

// NewPostgresRateLimitRepository создает новое хранилище в PostgreSQL
func NewClientRepo(db *pgxpool.Pool) Client {
	// Создаем таблицу, если она не существует
	// _, err := db.Exec(`
	// 	CREATE TABLE IF NOT EXISTS rate_limit_clients (
	// 		id VARCHAR(255) PRIMARY KEY,
	// 		capacity INT NOT NULL,
	// 		refill_rate FLOAT NOT NULL,
	// 		last_accessed TIMESTAMP NOT NULL,
	// 		created_at TIMESTAMP NOT NULL,
	// 		updated_at TIMESTAMP NOT NULL
	// 	)
	// `)
	// if err != nil {
	// 	panic(err)
	// }

	return &ClientRepo{db: db}
}

// CreateClient создает нового клиента
func (r *ClientRepo) CreateClient(ctx context.Context, client *entity.RateLimitClient) (int, error) {
	var id int
	query := "INSERT INTO rate_limit_clients (id, capacity, refill_rate) VALUES ($1, $2, $3) RETURNING id"

	row := r.db.QueryRow(ctx, query, client.ID, client.Capacity, client.RefillRate)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

// GetClient получает клиента по ID
func (r *ClientRepo) GetClient(ctx context.Context, clientID string) (*entity.RateLimitClient, error) {
	var client entity.RateLimitClient

	query := "SELECT id, capacity, refill_rate FROM rate_limit_clients WHERE id = $1"
	row := r.db.QueryRow(ctx, query, clientID)
	if err := row.Scan(&client.ID, &client.Capacity, &client.RefillRate); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("client not found")
		}
		return nil, err
	}
	return &client, nil
}

// UpdateClient обновляет клиента
func (r *ClientRepo) UpdateClient(ctx context.Context, client *entity.RateLimitClient) (int, error) {
	var id int
	query := "UPDATE rate_limit_clients SET capacity = $1, refill_rate = $2 WHERE id = $3 RETURNING id"
	row := r.db.QueryRow(ctx, query, client.Capacity, client.RefillRate, client.ID)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

// DeleteClient удаляет клиента
func (r *ClientRepo) DeleteClient(ctx context.Context, clientID string) (int, error) {
	var id int
	query := "DELETE FROM rate_limit_clients WHERE id = $1 RETURNING id"
	row := r.db.QueryRow(ctx, query, clientID)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

// ListClients возвращает список всех клиентов
func (r *ClientRepo) ListClients(ctx context.Context) ([]*entity.RateLimitClient, error) {
	rows, err := r.db.Query(ctx, "SELECT id, capacity, refill_rate FROM rate_limit_clients")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []*entity.RateLimitClient
	for rows.Next() {
		var client entity.RateLimitClient
		err := rows.Scan(
			&client.ID,
			&client.Capacity,
			&client.RefillRate,
		)
		if err != nil {
			return nil, err
		}
		clients = append(clients, &client)
	}
	return clients, rows.Err()
}
