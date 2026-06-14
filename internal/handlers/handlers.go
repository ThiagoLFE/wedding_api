package handlers

import (
	"encoding/json"
	"net/http"
	"wedding_api/internal/service"
)

type Handler struct {
	service *service.Service
}

func GetHandler(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Product
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {

	var request service.CreateProduct

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	product, err := h.service.AddProduct(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(product)
}
