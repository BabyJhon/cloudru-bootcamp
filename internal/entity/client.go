package entity

// RateLimitClient представляет клиента с настройками rate-limiting
type RateLimitClient struct {
	ID         string  `json:"client_id" db:"id"`
	Capacity   int     `json:"capacity" db:"capacity"`
	RefillRate float64 `json:"rate_per_sec" db:"refill_rate"`
}

// ClientList представляет список клиентов для API-запросов
type ClientList struct {
	Clients []RateLimitClient `json:"clients"`
	Total   int               `json:"total"`
}

// CreateClientRequest представляет запрос на создание клиента
type CreateClientRequest struct {
	ClientID   string  `json:"client_id" validate:"required"`
	Capacity   int     `json:"capacity" validate:"required,min=1"`
	RatePerSec float64 `json:"rate_per_sec" validate:"required,min=0.1"`
}

// UpdateClientRequest представляет запрос на обновление клиента
type UpdateClientRequest struct {
	Capacity   int     `json:"capacity" validate:"required,min=1"`
	RatePerSec float64 `json:"rate_per_sec" validate:"required,min=0.1"`
}
