package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
)

type mockBalancer struct {
	backends []*url.URL
	current  int
	mu       sync.Mutex
}

func newMockBalancer(backends []*url.URL) *mockBalancer {
	return &mockBalancer{
		backends: backends,
		current:  0,
	}
}

func (m *mockBalancer) Next() *url.URL {
	m.mu.Lock()
	defer m.mu.Unlock()
	backend := m.backends[m.current]
	m.current = (m.current + 1) % len(m.backends)
	return backend
}

func (m *mockBalancer) GetBackends() []*url.URL {
	return m.backends
}

type mockRateLimiter struct {
	allowed bool
}

func newMockRateLimiter(allowed bool) *mockRateLimiter {
	return &mockRateLimiter{allowed: allowed}
}

func (m *mockRateLimiter) IsAllowed(clientID string) bool {
	return m.allowed
}

func (m *mockRateLimiter) Stop() {}

type mockClientIdentifier struct {
	clientID string
}

func newMockClientIdentifier(clientID string) *mockClientIdentifier {
	return &mockClientIdentifier{clientID: clientID}
}

func (m *mockClientIdentifier) IdentifyClient(r *http.Request) string {
	return m.clientID
}

func (m *mockClientIdentifier) GetAPIKey(r *http.Request) string {
	return r.Header.Get("X-API-Key")
}

func (m *mockClientIdentifier) GetClientIP(r *http.Request) string {
	return r.RemoteAddr
}

func setupTestServer(statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte("test response"))
	}))
}

func TestProxyHandler_ConcurrentRequests(t *testing.T) {
	backend1 := setupTestServer(http.StatusOK)
	backend2 := setupTestServer(http.StatusOK)
	defer func() {
		backend1.Close()
		backend2.Close()
	}()

	backend1URL, _ := url.Parse(backend1.URL)
	backend2URL, _ := url.Parse(backend2.URL)

	balancer := newMockBalancer([]*url.URL{backend1URL, backend2URL})
	rateLimiter := newMockRateLimiter(true)
	clientIdentifier := newMockClientIdentifier("test-client")

	handler := NewProxyHandler(balancer, rateLimiter, clientIdentifier, 100)

	numRequests := 100
	var wg sync.WaitGroup
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()
			ctx := context.Background()
			req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
			}
		}()
	}

	wg.Wait()
}

func BenchmarkProxyHandler(b *testing.B) {
	backend := setupTestServer(http.StatusOK)
	defer backend.Close()

	backendURL, _ := url.Parse(backend.URL)

	balancer := newMockBalancer([]*url.URL{backendURL})
	rateLimiter := newMockRateLimiter(true)
	clientIdentifier := newMockClientIdentifier("test-client")

	handler := NewProxyHandler(balancer, rateLimiter, clientIdentifier, 100)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkProxyHandlerConcurrent(b *testing.B) {
	backend := setupTestServer(http.StatusOK)
	defer backend.Close()

	backendURL, _ := url.Parse(backend.URL)

	balancer := newMockBalancer([]*url.URL{backendURL})
	rateLimiter := newMockRateLimiter(true)
	clientIdentifier := newMockClientIdentifier("test-client")

	handler := NewProxyHandler(balancer, rateLimiter, clientIdentifier, 100)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			handler.ServeHTTP(w, req)
		}
	})
}

func BenchmarkProxyHandlerWithBody(b *testing.B) {
	backend := setupTestServer(http.StatusOK)
	defer backend.Close()

	backendURL, _ := url.Parse(backend.URL)

	balancer := newMockBalancer([]*url.URL{backendURL})
	rateLimiter := newMockRateLimiter(true)
	clientIdentifier := newMockClientIdentifier("test-client")

	handler := NewProxyHandler(balancer, rateLimiter, clientIdentifier, 100)

	body := []byte("test body")
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(w, req)
	}
}

func TestProxyHandler_RateLimit(t *testing.T) {
	backend := setupTestServer(http.StatusOK)
	defer func() {
		backend.Close()
	}()

	backendURL, _ := url.Parse(backend.URL)

	balancer := newMockBalancer([]*url.URL{backendURL})
	rateLimiter := newMockRateLimiter(false) // Запрещаем все запросы
	clientIdentifier := newMockClientIdentifier("test-client")

	handler := NewProxyHandler(balancer, rateLimiter, clientIdentifier, 100)

	ctx := context.Background()
	req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status code %d, got %d", http.StatusTooManyRequests, w.Code)
	}
}

func TestProxyHandler_Retry(t *testing.T) {
	backend1 := setupTestServer(http.StatusInternalServerError)
	backend2 := setupTestServer(http.StatusOK)
	defer func() {
		backend1.Close()
		backend2.Close()
	}()

	backend1URL, _ := url.Parse(backend1.URL)
	backend2URL, _ := url.Parse(backend2.URL)

	balancer := newMockBalancer([]*url.URL{backend1URL, backend2URL})
	rateLimiter := newMockRateLimiter(true)
	clientIdentifier := newMockClientIdentifier("test-client")

	handler := NewProxyHandler(balancer, rateLimiter, clientIdentifier, 100)

	ctx := context.Background()
	req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestProxyHandler_RequestBody(t *testing.T) {
	backend := setupTestServer(http.StatusOK)
	defer backend.Close()

	backendURL, _ := url.Parse(backend.URL)

	balancer := newMockBalancer([]*url.URL{backendURL})
	rateLimiter := newMockRateLimiter(true)
	clientIdentifier := newMockClientIdentifier("test-client")

	handler := NewProxyHandler(balancer, rateLimiter, clientIdentifier, 100)

	body := []byte("test body")
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}
