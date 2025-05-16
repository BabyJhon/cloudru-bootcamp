package service

import (
	"net"
	"net/http"
	"strings"
)

type ClientIdentifierService struct {
	prioritizeAPIKey bool
	ipPrefix         string
}

func NewClientIdentifierService(prioritizeAPIKey bool) *ClientIdentifierService {
	return &ClientIdentifierService{
		prioritizeAPIKey: prioritizeAPIKey,
		ipPrefix:         "ip:",
	}
}

// определяет ID клиента из запроса
func (s *ClientIdentifierService) IdentifyClient(r *http.Request) string {
	//fmt.Println("=========================")

	if s.prioritizeAPIKey {
		if id := s.GetAPIKey(r); id != "" {
			return id
		}
	}

	return s.ipPrefix + s.GetClientIP(r)
}

func (s *ClientIdentifierService) GetAPIKey(r *http.Request) string {
	// проверяем заголовок, авторизационный токен и параметр запроса
	apiKey := r.Header.Get("X-API-Key")
	if apiKey != "" {
		return apiKey
	}

	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	if key := r.URL.Query().Get("api_key"); key != "" {
		return key
	}

	return ""
}

func (s *ClientIdentifierService) GetClientIP(r *http.Request) string {
	// проверяем заголовки X-Forwarded-For и X-Real-IP и в крайнем случае используем RemoteAddr
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
