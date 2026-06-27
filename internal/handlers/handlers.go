package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"wedding_api/ent"
	"wedding_api/internal/service"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service ServiceAPI
}

type confirmationPresenceResponse struct {
	ID          int    `json:"id"`
	Fullname    string `json:"fullname"`
	PhotoBase64 string `json:"photo_base64"`
	IsConfirmed bool   `json:"is_confirmed"`
}

type ServiceAPI interface {
	AddProduct(ctx context.Context, product service.ProductRequest) (*ent.Product, error)
	ProductsList(ctx context.Context) ([]*ent.Product, error)
	GetProduct(ctx context.Context, id int) (*ent.Product, error)
	UpdateProduct(ctx context.Context, newProduct service.UpdateProduct) (*ent.Product, error)
	DeleteProduct(ctx context.Context, id int) error
	AddConfirmationPresence(ctx context.Context, presence service.ConfirmationPresenceRequest) (*ent.ConfirmationPresence, error)
	ConfirmationPresencesList(ctx context.Context) ([]*ent.ConfirmationPresence, error)
	GetConfirmationPresence(ctx context.Context, id int) (*ent.ConfirmationPresence, error)
	UpdateConfirmationPresence(ctx context.Context, newPresence service.UpdateConfirmationPresence) (*ent.ConfirmationPresence, error)
	ConfirmPresence(ctx context.Context, id int) (*ent.ConfirmationPresence, error)
	CancelPresence(ctx context.Context, id int) (*ent.ConfirmationPresence, error)
	DeleteConfirmationPresence(ctx context.Context, id int) error
}

func NewHandler(service ServiceAPI) *Handler {
	return &Handler{
		service: service,
	}
}

// Config
func (h *Handler) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (h *Handler) NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not found", http.StatusNotFound)
}

// Products
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {

	var request service.ProductRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	product, err := h.service.AddProduct(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.ProductsList(r.Context())
	if err != nil {
		http.Error(w, fmt.Errorf("failed to get products: %w", err).Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(products)
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id, err := getURLIntParam(r, "id")

	if err != nil {
		http.Error(w, fmt.Errorf("failed to read url param: %w", err).Error(), http.StatusBadRequest)
		return
	}

	product, err := h.service.GetProduct(r.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(product)
}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := getURLIntParam(r, "id")

	if err != nil {
		http.Error(w, fmt.Errorf("failed to read url param: %w", err).Error(), http.StatusBadRequest)
		return
	}

	var response service.ProductRequest

	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		http.Error(w, fmt.Errorf("invalid json: %w", err).Error(), http.StatusBadRequest)
		return
	}

	product, err := h.service.UpdateProduct(r.Context(), service.UpdateProduct{ID: id, ProductRequest: response})
	if err != nil {
		if ent.IsNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(product)
}

func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := getURLIntParam(r, "id")
	if err != nil {
		http.Error(w, fmt.Errorf("failed to read url param: %w", err).Error(), http.StatusBadRequest)
		return
	}
	if err := h.service.DeleteProduct(r.Context(), id); err != nil {
		if ent.IsNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

// Confirmation presences
func (h *Handler) CreateConfirmationPresence(w http.ResponseWriter, r *http.Request) {
	var request service.ConfirmationPresenceRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	presence, err := h.service.AddConfirmationPresence(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toConfirmationPresenceResponse(presence))
}

func (h *Handler) GetConfirmationPresences(w http.ResponseWriter, r *http.Request) {
	presences, err := h.service.ConfirmationPresencesList(r.Context())
	if err != nil {
		http.Error(w, fmt.Errorf("failed to get confirmation presences: %w", err).Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(toConfirmationPresenceResponses(presences))
}

func (h *Handler) GetConfirmationPresence(w http.ResponseWriter, r *http.Request) {
	id, err := getURLIntParam(r, "id")
	if err != nil {
		http.Error(w, fmt.Errorf("failed to read url param: %w", err).Error(), http.StatusBadRequest)
		return
	}

	presence, err := h.service.GetConfirmationPresence(r.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(toConfirmationPresenceResponse(presence))
}

func (h *Handler) UpdateConfirmationPresence(w http.ResponseWriter, r *http.Request) {
	id, err := getURLIntParam(r, "id")
	if err != nil {
		http.Error(w, fmt.Errorf("failed to read url param: %w", err).Error(), http.StatusBadRequest)
		return
	}

	var request service.ConfirmationPresenceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Errorf("invalid json: %w", err).Error(), http.StatusBadRequest)
		return
	}

	presence, err := h.service.UpdateConfirmationPresence(r.Context(), service.UpdateConfirmationPresence{ID: id, ConfirmationPresenceRequest: request})
	if err != nil {
		if ent.IsNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(toConfirmationPresenceResponse(presence))
}

func (h *Handler) ConfirmPresence(w http.ResponseWriter, r *http.Request) {
	h.updatePresenceConfirmation(w, r, true)
}

func (h *Handler) CancelPresence(w http.ResponseWriter, r *http.Request) {
	h.updatePresenceConfirmation(w, r, false)
}

func (h *Handler) DeleteConfirmationPresence(w http.ResponseWriter, r *http.Request) {
	id, err := getURLIntParam(r, "id")
	if err != nil {
		http.Error(w, fmt.Errorf("failed to read url param: %w", err).Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteConfirmationPresence(r.Context(), id); err != nil {
		if ent.IsNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *Handler) updatePresenceConfirmation(w http.ResponseWriter, r *http.Request, confirmed bool) {
	id, err := getURLIntParam(r, "id")
	if err != nil {
		http.Error(w, fmt.Errorf("failed to read url param: %w", err).Error(), http.StatusBadRequest)
		return
	}

	var presence *ent.ConfirmationPresence
	if confirmed {
		presence, err = h.service.ConfirmPresence(r.Context(), id)
	} else {
		presence, err = h.service.CancelPresence(r.Context(), id)
	}
	if err != nil {
		if ent.IsNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(toConfirmationPresenceResponse(presence))
}

func toConfirmationPresenceResponse(presence *ent.ConfirmationPresence) confirmationPresenceResponse {
	return confirmationPresenceResponse{
		ID:          presence.ID,
		Fullname:    presence.Fullname,
		PhotoBase64: presence.PhotoBase64,
		IsConfirmed: presence.IsConfirmed,
	}
}

func toConfirmationPresenceResponses(presences []*ent.ConfirmationPresence) []confirmationPresenceResponse {
	result := make([]confirmationPresenceResponse, 0, len(presences))
	for _, presence := range presences {
		result = append(result, toConfirmationPresenceResponse(presence))
	}
	return result
}

// Helpers
func getURLIntParam(r *http.Request, labelParam string) (int, error) {
	intParam, err := strconv.Atoi(chi.URLParam(r, labelParam))
	return intParam, err
}
