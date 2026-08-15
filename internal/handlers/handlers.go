package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"wedding_api/ent"
	"wedding_api/internal/auth"
	"wedding_api/internal/service"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service ServiceAPI
}

type confirmationPresenceResponse struct {
	ID          int    `json:"id"`
	FamilyID    int    `json:"family_id"`
	Fullname    string `json:"fullname"`
	PhotoBase64 string `json:"photo_base64"`
	IsConfirmed bool   `json:"is_confirmed"`
}

type familyResponse struct {
	ID        int                            `json:"id"`
	Name      string                         `json:"name"`
	Presences []confirmationPresenceResponse `json:"presences,omitempty"`
}

type sessionResponse struct {
	Role     string `json:"role"`
	UserID   *int   `json:"user_id,omitempty"`
	FamilyID *int   `json:"family_id,omitempty"`
}

type reserveProductRequest struct {
	ReservedBy string `json:"reserved_by"`
}

type ServiceAPI interface {
	AddProduct(context.Context, service.ProductRequest) (*ent.Product, error)
	ProductsList(context.Context) ([]*ent.Product, error)
	GetProduct(context.Context, int) (*ent.Product, error)
	UpdateProduct(context.Context, service.UpdateProduct) (*ent.Product, error)
	ReserveProduct(context.Context, int, string) (*ent.Product, error)
	DeleteProduct(context.Context, int) error

	AddConfirmationPresence(context.Context, service.ConfirmationPresenceRequest) (*ent.ConfirmationPresence, error)
	ConfirmationPresencesList(context.Context) ([]*ent.ConfirmationPresence, error)
	GetConfirmationPresence(context.Context, int) (*ent.ConfirmationPresence, error)
	UpdateConfirmationPresence(context.Context, service.UpdateConfirmationPresence) (*ent.ConfirmationPresence, error)
	ConfirmPresence(context.Context, int) (*ent.ConfirmationPresence, error)
	CancelPresence(context.Context, int) (*ent.ConfirmationPresence, error)
	DeleteConfirmationPresence(context.Context, int) error

	AdminLogin(context.Context, service.AdminLoginRequest) (*service.LoginResult, error)
	ExchangeFamilyToken(context.Context, string) (*service.LoginResult, error)
	Logout(context.Context, string) error

	CreateFamily(context.Context, service.FamilyRequest) (*ent.Family, error)
	FamiliesList(context.Context) ([]*ent.Family, error)
	GetFamily(context.Context, int) (*ent.Family, error)
	GetMyFamily(context.Context) (*ent.Family, error)
	UpdateFamily(context.Context, int, service.FamilyRequest) (*ent.Family, error)
	DeleteFamily(context.Context, int) error
	CreateFamilyAccessLink(context.Context, int) (string, error)
	RevokeFamilyAccessLink(context.Context, int) error
}

func NewHandler(s ServiceAPI) *Handler {
	return &Handler{service: s}
}

// AuthenticateSession lets the router authenticate requests without exposing
// the concrete service implementation to the routing package.
func (h *Handler) AuthenticateSession(ctx context.Context, token string) (*auth.Principal, error) {
	authenticator, ok := h.service.(auth.SessionAuthenticator)
	if !ok {
		return nil, auth.ErrInvalidToken
	}
	return authenticator.AuthenticateSession(ctx, token)
}

func (h *Handler) MethodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (h *Handler) NotFound(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "not found", http.StatusNotFound)
}

// Authentication
func (h *Handler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	var request service.AdminLoginRequest
	if err := decodeJSON(r, &request); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	result, err := h.service.AdminLogin(r.Context(), request)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	auth.SetSessionCookie(w, result.Token, result.ExpiresAt, secureCookies())
	writeJSON(w, http.StatusOK, toSessionResponse(result.Principal))
}

func (h *Handler) ExchangeFamilyToken(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Token string `json:"token"`
	}
	if err := decodeJSON(r, &request); err != nil || request.Token == "" {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	result, err := h.service.ExchangeFamilyToken(r.Context(), request.Token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	auth.SetSessionCookie(w, result.Token, result.ExpiresAt, secureCookies())
	writeJSON(w, http.StatusOK, toSessionResponse(result.Principal))
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(auth.SessionCookieName); err == nil {
		_ = h.service.Logout(r.Context(), cookie.Value)
	}
	auth.ClearSessionCookie(w, secureCookies())
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	principal, err := auth.RequirePrincipal(r.Context())
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, toSessionResponse(principal))
}

// Families
func (h *Handler) CreateFamily(w http.ResponseWriter, r *http.Request) {
	var request service.FamilyRequest
	if err := decodeJSON(r, &request); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	family, err := h.service.CreateFamily(r.Context(), request)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toFamilyResponse(family))
}

func (h *Handler) GetFamilies(w http.ResponseWriter, r *http.Request) {
	families, err := h.service.FamiliesList(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	result := make([]familyResponse, 0, len(families))
	for _, family := range families {
		result = append(result, toFamilyResponse(family))
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetFamily(w http.ResponseWriter, r *http.Request) {
	id, err := getURLIntParam(r, "id")
	if err != nil {
		http.Error(w, "invalid family id", http.StatusBadRequest)
		return
	}
	family, err := h.service.GetFamily(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toFamilyResponse(family))
}

func (h *Handler) GetMyFamily(w http.ResponseWriter, r *http.Request) {
	family, err := h.service.GetMyFamily(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toFamilyResponse(family))
}

func (h *Handler) UpdateFamily(w http.ResponseWriter, r *http.Request) {
	id, err := getURLIntParam(r, "id")
	if err != nil {
		http.Error(w, "invalid family id", http.StatusBadRequest)
		return
	}
	var request service.FamilyRequest
	if err := decodeJSON(r, &request); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	family, err := h.service.UpdateFamily(r.Context(), id, request)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toFamilyResponse(family))
}

func (h *Handler) DeleteFamily(w http.ResponseWriter, r *http.Request) {
	id, err := getURLIntParam(r, "id")
	if err != nil {
		http.Error(w, "invalid family id", http.StatusBadRequest)
		return
	}
	if err := h.service.DeleteFamily(r.Context(), id); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CreateFamilyAccessLink(w http.ResponseWriter, r *http.Request) {
	id, err := getURLIntParam(r, "id")
	if err != nil {
		http.Error(w, "invalid family id", http.StatusBadRequest)
		return
	}
	token, err := h.service.CreateFamilyAccessLink(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	baseURL := os.Getenv("FAMILY_ACCESS_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000/access"
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"family_id": id,
		"url":       baseURL + "?token=" + url.QueryEscape(token),
	})
}

func (h *Handler) RevokeFamilyAccessLink(w http.ResponseWriter, r *http.Request) {
	id, err := getURLIntParam(r, "id")
	if err != nil {
		http.Error(w, "invalid family id", http.StatusBadRequest)
		return
	}
	if err := h.service.RevokeFamilyAccessLink(r.Context(), id); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Products
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var request service.ProductRequest
	if err := decodeJSON(r, &request); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	product, err := h.service.AddProduct(r.Context(), request)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, product)
}

func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.ProductsList(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, products)
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id, err := getURLIntParam(r, "id")
	if err != nil {
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}
	product, err := h.service.GetProduct(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := getURLIntParam(r, "id")
	if err != nil {
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}
	var request service.ProductRequest
	if err := decodeJSON(r, &request); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	product, err := h.service.UpdateProduct(r.Context(), service.UpdateProduct{ID: id, ProductRequest: request})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func (h *Handler) ReserveProduct(w http.ResponseWriter, r *http.Request) {
	id, err := getURLIntParam(r, "id")
	if err != nil {
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}
	var request reserveProductRequest
	if err := decodeJSON(r, &request); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	product, err := h.service.ReserveProduct(r.Context(), id, request.ReservedBy)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := getURLIntParam(r, "id")
	if err != nil {
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}
	if err := h.service.DeleteProduct(r.Context(), id); err != nil {
		writeServiceError(w, err)
		return
	}
}

// Confirmation presences
func (h *Handler) CreateConfirmationPresence(w http.ResponseWriter, r *http.Request) {
	var request service.ConfirmationPresenceRequest
	if err := decodeJSON(r, &request); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	presence, err := h.service.AddConfirmationPresence(r.Context(), request)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toConfirmationPresenceResponse(presence))
}

func (h *Handler) GetConfirmationPresences(w http.ResponseWriter, r *http.Request) {
	presences, err := h.service.ConfirmationPresencesList(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toConfirmationPresenceResponses(presences))
}

func (h *Handler) GetConfirmationPresence(w http.ResponseWriter, r *http.Request) {
	id, err := getURLIntParam(r, "id")
	if err != nil {
		http.Error(w, "invalid presence id", http.StatusBadRequest)
		return
	}
	presence, err := h.service.GetConfirmationPresence(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toConfirmationPresenceResponse(presence))
}

func (h *Handler) UpdateConfirmationPresence(w http.ResponseWriter, r *http.Request) {
	id, err := getURLIntParam(r, "id")
	if err != nil {
		http.Error(w, "invalid presence id", http.StatusBadRequest)
		return
	}
	var request service.ConfirmationPresenceRequest
	if err := decodeJSON(r, &request); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	presence, err := h.service.UpdateConfirmationPresence(r.Context(), service.UpdateConfirmationPresence{ID: id, ConfirmationPresenceRequest: request})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toConfirmationPresenceResponse(presence))
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
		http.Error(w, "invalid presence id", http.StatusBadRequest)
		return
	}
	if err := h.service.DeleteConfirmationPresence(r.Context(), id); err != nil {
		writeServiceError(w, err)
		return
	}
}

func (h *Handler) updatePresenceConfirmation(w http.ResponseWriter, r *http.Request, confirmed bool) {
	id, err := getURLIntParam(r, "id")
	if err != nil {
		http.Error(w, "invalid presence id", http.StatusBadRequest)
		return
	}
	var presence *ent.ConfirmationPresence
	if confirmed {
		presence, err = h.service.ConfirmPresence(r.Context(), id)
	} else {
		presence, err = h.service.CancelPresence(r.Context(), id)
	}
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toConfirmationPresenceResponse(presence))
}

func toSessionResponse(principal *auth.Principal) sessionResponse {
	return sessionResponse{Role: principal.Role, UserID: principal.UserID, FamilyID: principal.FamilyID}
}

func toConfirmationPresenceResponse(presence *ent.ConfirmationPresence) confirmationPresenceResponse {
	familyID := 0
	if presence.Edges.Family != nil {
		familyID = presence.Edges.Family.ID
	}
	return confirmationPresenceResponse{ID: presence.ID, FamilyID: familyID, Fullname: presence.Fullname, PhotoBase64: presence.PhotoBase64, IsConfirmed: presence.IsConfirmed}
}

func toConfirmationPresenceResponses(presences []*ent.ConfirmationPresence) []confirmationPresenceResponse {
	result := make([]confirmationPresenceResponse, 0, len(presences))
	for _, presence := range presences {
		result = append(result, toConfirmationPresenceResponse(presence))
	}
	return result
}

func toFamilyResponse(family *ent.Family) familyResponse {
	result := familyResponse{ID: family.ID, Name: family.Name}
	if len(family.Edges.Presences) > 0 {
		result.Presences = toConfirmationPresenceResponses(family.Edges.Presences)
	}
	return result
}

func decodeJSON(r *http.Request, value any) error {
	return json.NewDecoder(r.Body).Decode(value)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeServiceError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	switch {
	case errors.Is(err, auth.ErrUnauthenticated):
		status = http.StatusUnauthorized
	case errors.Is(err, auth.ErrForbidden):
		status = http.StatusForbidden
	case ent.IsNotFound(err):
		status = http.StatusNotFound
	}
	http.Error(w, fmt.Sprintf("%v", err), status)
}

func secureCookies() bool {
	return os.Getenv("COOKIE_SECURE") == "true"
}

func getURLIntParam(r *http.Request, labelParam string) (int, error) {
	return strconv.Atoi(chi.URLParam(r, labelParam))
}
