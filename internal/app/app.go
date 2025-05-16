package app

import (
	"log"
	"net/http"

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

	log.Printf("Starting proxy server on :%s", cfg.ProxyPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ProxyPort, router))
}
