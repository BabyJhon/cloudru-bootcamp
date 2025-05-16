package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BabyJhon/cloudru-bootcamp/configs"
	"github.com/BabyJhon/cloudru-bootcamp/internal/handler"
	"github.com/BabyJhon/cloudru-bootcamp/internal/service"
	"github.com/gorilla/mux"
)

func Run() {
	cfg := configs.Load()

	backends := configs.NewBackends(cfg.BackendURLs)
	backendURLs, err := backends.GetBackends()
	if err != nil {
		log.Fatal("Error loading backends:", err)
	}

	// инициализация всех сервисов
	services := service.NewService(backendURLs)

	// Создаём роутер
	router := mux.NewRouter()

	// API для управления лимитами
	rateLimitHandler := handler.NewRateLimitHandler(services.ClientService)
	rateLimitHandler.RegisterRoutes(router)

	// Прокси-обработчик
	proxyHandler := handler.NewProxyHandler(
		services.Balancer,
		services.RateLimiter,
		services.ClientIdentifier,
		10, // concurrentLimit
	)

	// Все остальные запросы идут через прокси
	router.PathPrefix("/").Handler(proxyHandler)

	// Создаем HTTP сервер
	srv := &http.Server{
		Addr:    ":" + cfg.ProxyPort,
		Handler: router,
	}

	// Канал для получения сигналов завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запускаем сервер в горутине
	go func() {
		log.Printf("Starting proxy server on :%s", cfg.ProxyPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Ждем сигнал завершения
	<-stop
	log.Println("Shutting down server...")

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Останавливаем сервер
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	// Останавливаем все сервисы
	services.RateLimiter.Stop()
	log.Println("Server stopped gracefully")
}