package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"wedding_api/ent"
	"wedding_api/internal/service"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
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

	if err := json.NewDecoder(r.Body).Decode(response); err != nil {
		http.Error(w, fmt.Errorf("invalid json: %w", err).Error(), http.StatusBadRequest)
		return
	}

	h.service.UpdateProduct(r.Context(), service.UpdateProduct{ID: id, ProductRequest: response})
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

// Helpers
func getURLIntParam(r *http.Request, labelParam string) (int, error) {
	intParam, err := strconv.Atoi(chi.URLParam(r, labelParam))
	return intParam, err
}
