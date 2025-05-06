package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"http-load-balancer/models"
	"http-load-balancer/repository"
)

type ClientHandler struct {
	userRepo repository.UserRepository
}

func NewClientHandler(userRepo repository.UserRepository) *ClientHandler {
	return &ClientHandler{userRepo: userRepo}
}

func (h *ClientHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var clientReq struct {
		Capacity   int `json:"capacity"`
		RatePerSec int `json:"rate_per_sec"`
	}

	if err := json.NewDecoder(r.Body).Decode(&clientReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if clientReq.Capacity <= 0 {
		http.Error(w, "capacity must be positive", http.StatusBadRequest)
		return
	}
	if clientReq.RatePerSec <= 0 {
		http.Error(w, "rate_per_sec must be positive", http.StatusBadRequest)
		return
	}

	reqUser := &models.User{
		Capacity:   clientReq.Capacity,
		RatePerSec: clientReq.RatePerSec,
	}

	user, err := h.userRepo.Create(reqUser)
	if err != nil {
		http.Error(w, "Failed to create client", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "success",
		"client_id": user.ID,
	})
}

func (h *ClientHandler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	rawClientID := r.PathValue("client_id")
	if rawClientID == "" {
		http.Error(w, "Client ID is required", http.StatusBadRequest)
		return
	}

	clientID, err := strconv.ParseUint(rawClientID, 10, 64)
	if err != nil {
		http.Error(w, "Invalid client_id", http.StatusBadRequest)
	}

	if err := h.userRepo.Delete(clientID); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			http.Error(w, "Client not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete client", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "success",
		"client_id": clientID,
		"message":   "Client deleted successfully",
	})
}

func (h *ClientHandler) UpdateClientParams(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	rawClientID := r.PathValue("client_id")
	if rawClientID == "" {
		http.Error(w, "Client ID is required", http.StatusBadRequest)
	}
	clientID, err := strconv.ParseUint(rawClientID, 10, 64)
	if err != nil {
		http.Error(w, "Invalid client_id", http.StatusBadRequest)
	}

	var updateReq struct {
		Capacity   int `json:"capacity,omitempty"`
		RatePerSec int `json:"rate_per_sec,omitempty"`
		Tokens     int `json:"tokens,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if updateReq.Capacity <= 0 {
		http.Error(w, "Capacity must be positive", http.StatusBadRequest)
		return
	}
	if updateReq.RatePerSec <= 0 {
		http.Error(w, "Rate_per_sec must be positive", http.StatusBadRequest)
		return
	}
	if updateReq.Tokens <= 0 {
		http.Error(w, "Tokens must be positive", http.StatusBadRequest)
	}

	existingUser, err := h.userRepo.GetByID(clientID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	existingUser.Capacity = updateReq.Capacity
	existingUser.RatePerSec = updateReq.RatePerSec
	existingUser.Tokens = updateReq.Tokens

	if err := h.userRepo.Update(&existingUser); err != nil {
		http.Error(w, "Failed to update client", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "success",
		"client_id":    clientID,
		"capacity":     existingUser.Capacity,
		"rate_per_sec": existingUser.RatePerSec,
		"tokens":       existingUser.Tokens,
	})
}
