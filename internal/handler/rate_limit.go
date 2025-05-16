package handler

import (
	"encoding/json"
	"net/http"

	"github.com/BabyJhon/cloudru-bootcamp/internal/entity"
	"github.com/BabyJhon/cloudru-bootcamp/internal/service"
	"github.com/gorilla/mux"
)

type RateLimitHandler struct {
	clientService *service.ClientService
}

func NewRateLimitHandler(clientService *service.ClientService) *RateLimitHandler {
	return &RateLimitHandler{
		clientService: clientService,
	}
}

func (h *RateLimitHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/ratelimit/clients", h.ListClients).Methods("GET")
	router.HandleFunc("/api/ratelimit/clients", h.CreateClient).Methods("POST")
	router.HandleFunc("/api/ratelimit/clients/{clientID}", h.GetClient).Methods("GET")
	router.HandleFunc("/api/ratelimit/clients/{clientID}", h.UpdateClient).Methods("PUT")
	router.HandleFunc("/api/ratelimit/clients/{clientID}", h.DeleteClient).Methods("DELETE")
	router.HandleFunc("/api/ratelimit/clients/{clientID}/tokens", h.GetClientTokens).Methods("GET")
}

func (h *RateLimitHandler) ListClients(w http.ResponseWriter, r *http.Request) {
	response := h.clientService.ListClients()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *RateLimitHandler) GetClient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["clientID"]

	client, err := h.clientService.GetClient(clientID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(entity.ErrorResponse{
			Code:    http.StatusNotFound,
			Message: "Client not found",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(client)
}

func (h *RateLimitHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	var req entity.CreateClientRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(entity.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}

	err := h.clientService.CreateClient(&req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(entity.ErrorResponse{
			Code:    http.StatusConflict,
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *RateLimitHandler) UpdateClient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["clientID"]

	var req entity.UpdateClientRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(entity.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}

	err := h.clientService.UpdateClient(clientID, &req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(entity.ErrorResponse{
			Code:    http.StatusNotFound,
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *RateLimitHandler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["clientID"]

	err := h.clientService.DeleteClient(clientID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(entity.ErrorResponse{
			Code:    http.StatusNotFound,
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// GetClientTokens возвращает информацию о количестве оставшихся токенов
func (h *RateLimitHandler) GetClientTokens(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["clientID"]

	tokens, err := h.clientService.GetTokensRemaining(clientID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(entity.ErrorResponse{
			Code:    http.StatusNotFound,
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"client_id": clientID,
		"tokens":    tokens,
	})
}
