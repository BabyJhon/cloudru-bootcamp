package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/BabyJhon/cloudru-bootcamp/internal/service"
)

type ProxyHandler struct {
	balancer         service.Balancer
	rateLimiter      service.RateLimiterService
	clientIdentifier service.ClientIdentifier
	proxy            *httputil.ReverseProxy
	maxRetries       int
	bufferPool       *sync.Pool // пул буферов для тела запроса
	semaphore        chan struct{}
	// защита от повторной записи заголовков
	responseWritten sync.Map
}

type contextKey string

const (
	retriesKey        contextKey = "retries"
	originalBodyKey   contextKey = "originalBody"
	currentBackendKey contextKey = "currentBackend"
	startTimeKey      contextKey = "startTime"
	requestIDKey      contextKey = "requestID"
)

// обертка для отслеживания записи заголовков
type responseWriterWrapper struct {
	http.ResponseWriter
	written      atomic.Bool
	requestID    string
	proxyHandler *ProxyHandler
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	if w.written.Load() {
		log.Printf("[WARNING][%s] Prevented duplicate WriteHeader call", w.requestID)
		return
	}
	w.written.Store(true)
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	// если заголовки еще не были записаны, автоматически записываем 200
	if !w.written.Load() {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

func NewProxyHandler(balancer service.Balancer, rateLimiter service.RateLimiterService,
	clientIdentifier service.ClientIdentifier, concurrentLimit int) *ProxyHandler {
	if concurrentLimit <= 0 {
		concurrentLimit = 100
	}

	ph := &ProxyHandler{
		balancer:         balancer,
		rateLimiter:      rateLimiter,
		clientIdentifier: clientIdentifier,
		maxRetries:       len(balancer.GetBackends()),
		bufferPool:       &sync.Pool{New: func() interface{} { return new(bytes.Buffer) }},
		semaphore:        make(chan struct{}, concurrentLimit),
	}

	director := func(req *http.Request) {
		requestID, _ := req.Context().Value(requestIDKey).(string)

		// сохраняем тело запроса при первом обращении для возможных повторных попыток
		if req.Context().Value(originalBodyKey) == nil && req.Body != nil {
			buf := ph.bufferPool.Get().(*bytes.Buffer)
			buf.Reset()
			defer ph.bufferPool.Put(buf)

			_, err := io.Copy(buf, req.Body)
			req.Body.Close()
			if err != nil {
				log.Printf("[ERROR][%s] Failed to read request body: %v", requestID, err)
				return
			}

			bodyBytes := make([]byte, buf.Len())
			copy(bodyBytes, buf.Bytes())
			ctx := context.WithValue(req.Context(), originalBodyKey, bodyBytes)
			*req = *req.WithContext(ctx)

			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		target := ph.balancer.Next()

		ctx := context.WithValue(req.Context(), currentBackendKey, target)
		*req = *req.WithContext(ctx)

		retries, _ := req.Context().Value(retriesKey).(int)
		if retries > 0 {
			log.Printf("[RETRY][%s] %d/%d Routing request to: %s",
				requestID, retries, ph.maxRetries-1, target.String())
		} else {
			log.Printf("[ROUTING][%s] Request forwarded to backend: %s",
				requestID, target.String())
		}

		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		req.Host = target.Host
	}

	ph.proxy = &httputil.ReverseProxy{
		Director:       director,
		ErrorHandler:   ph.errorHandler,
		ModifyResponse: ph.modifyResponse,
		Transport: &http.Transport{
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       30 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
	}

	return ph
}

func (h *ProxyHandler) modifyResponse(resp *http.Response) error {
	requestID, _ := resp.Request.Context().Value(requestIDKey).(string)

	// если ответ успешный, логируем информацию
	if resp.StatusCode < 500 {
		currentBackend, _ := resp.Request.Context().Value(currentBackendKey).(*url.URL)
		startTime, _ := resp.Request.Context().Value(startTimeKey).(time.Time)
		duration := time.Since(startTime)

		log.Printf("[SUCCESS][%s] Backend %s returned status %d in %v",
			requestID, currentBackend, resp.StatusCode, duration)
		return nil
	}

	currentBackend, _ := resp.Request.Context().Value(currentBackendKey).(*url.URL)
	return fmt.Errorf("backend server %s returned error status: %d", currentBackend, resp.StatusCode)
}

func (h *ProxyHandler) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	requestID, _ := r.Context().Value(requestIDKey).(string)
	retries, _ := r.Context().Value(retriesKey).(int)
	currentBackend, _ := r.Context().Value(currentBackendKey).(*url.URL)

	// проверяем, не превышен ли лимит попыток
	if retries >= h.maxRetries-1 {
		log.Printf("[FAILED][%s] Max retries reached (%d/%d). Last error from %s: %v",
			requestID, retries+1, h.maxRetries, currentBackend, err)

		startTime, _ := r.Context().Value(startTimeKey).(time.Time)
		duration := time.Since(startTime)
		log.Printf("[REQUEST FAILED][%s] %s %s -> FAILED after %v with %d attempts",
			requestID, r.Method, r.URL.Path, duration, retries+1)

		// проверяем, был ли уже отправлен ответ
		wrapper, ok := w.(*responseWriterWrapper)
		if ok && !wrapper.written.Load() {
			// Проверяем, является ли ошибка таймаутом контекста
			if err == context.DeadlineExceeded || errors.Is(err, context.DeadlineExceeded) {
				w.WriteHeader(http.StatusGatewayTimeout)
				w.Write([]byte("Gateway timeout"))
			} else {
				w.WriteHeader(http.StatusBadGateway)
				w.Write([]byte("All backend servers failed to process the request"))
			}
		}

		// освобождаем семафор
		select {
		case <-h.semaphore:
		default:
		}

		return
	}

	newCtx := context.WithValue(r.Context(), retriesKey, retries+1)

	// восстанавливаем тело запроса
	if bodyBytes, ok := r.Context().Value(originalBodyKey).([]byte); ok && len(bodyBytes) > 0 {
		r = r.Clone(newCtx)
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	} else {
		r = r.WithContext(newCtx)
	}

	log.Printf("[ERROR][%s] Backend %s returned error: %v. Retrying request (attempt %d/%d)...",
		requestID, currentBackend, err, retries+1, h.maxRetries)

	// выполняем следующую попытку синхронно, а не в горутине
	h.proxy.ServeHTTP(w, r)
}

func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := fmt.Sprintf("%d-%s", time.Now().UnixNano(), r.RemoteAddr)
	clientID := h.clientIdentifier.IdentifyClient(r)

	// вызов rate limiter
	if !h.rateLimiter.IsAllowed(clientID) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": "Rate limit exceeded"}`))
		log.Printf("[RATE LIMIT][%s] Request from client %s was rejected due to rate limit", requestID, clientID)
		return
	}

	// логируем входящий запрос
	startTime := time.Now()
	remoteIP := r.RemoteAddr
	userAgent := r.Header.Get("User-Agent")
	xForwardedFor := r.Header.Get("X-Forwarded-For")

	log.Printf("[REQUEST][%s] %s %s from %s, User-Agent: %s, X-Forwarded-For: %s",
		requestID, r.Method, r.URL.String(), remoteIP, userAgent, xForwardedFor)

	// Используем контекст из запроса, если он уже содержит таймаут
	ctx := r.Context()
	if _, ok := ctx.Deadline(); !ok {
		// Если в контексте нет таймаута, создаем новый с таймаутом по умолчанию
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
	}

	// контекст запроса
	ctx = context.WithValue(ctx, retriesKey, 0)
	ctx = context.WithValue(ctx, startTimeKey, startTime)
	ctx = context.WithValue(ctx, requestIDKey, requestID)
	r = r.WithContext(ctx)

	// ограничиваем количество одновременных запросов
	select {
	case h.semaphore <- struct{}{}:
		defer func() { <-h.semaphore }()
	default:
		log.Printf("[OVERLOAD][%s] Request rejected due to server overload", requestID)
		http.Error(w, "Server is overloaded", http.StatusServiceUnavailable)
		return
	}

	// оборачиваем ResponseWriter для отслеживания записи заголовков
	wrappedWriter := &responseWriterWrapper{
		ResponseWriter: w,
		requestID:      requestID,
		proxyHandler:   h,
	}

	h.proxy.ServeHTTP(wrappedWriter, r)
}
