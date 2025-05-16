package middleware

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/BabyJhon/cloudru-bootcamp/internal/entity"
	"github.com/BabyJhon/cloudru-bootcamp/internal/service"
)

func RateLimitMiddleware(rateLimiter service.RateLimiterService, identifier service.ClientIdentifier) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientID := identifier.IdentifyClient(r)

			allowed := rateLimiter.IsAllowed(clientID)

			if !allowed {
				// устанавливаем заголовки для ответа 429 Too Many Requests
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "1") // рекомендация подождать 1 секунду

				w.WriteHeader(http.StatusTooManyRequests)
				response := entity.ErrorResponse{
					Code:    http.StatusTooManyRequests,
					Message: "Rate limit exceeded. Please try again later.",
				}
				json.NewEncoder(w).Encode(response)
				return
			}

			// запрос прошел проверку, передаем его дальше
			next.ServeHTTP(w, r)
		})
	}
}

func RateLimitHeaders(clientService *service.ClientService, identifier service.ClientIdentifier) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientID := identifier.IdentifyClient(r)

			client, err := clientService.GetClient(clientID)
			if err == nil {
				// добавляем заголовки о лимитах
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(client.Capacity))

				// получаем оставшиеся токены
				tokens, err := clientService.GetTokensRemaining(clientID)
				if err == nil {
					w.Header().Set("X-RateLimit-Remaining", strconv.FormatFloat(tokens, 'f', 1, 64))
				}
			}

			// обрабатываем запрос (передаем дальше даже если клиент не найден)
			next.ServeHTTP(w, r)
		})
	}
}
