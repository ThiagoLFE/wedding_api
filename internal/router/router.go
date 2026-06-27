package router

import (
	"net/http"
	"wedding_api/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(handler *handlers.Handler) http.Handler {

	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:5173",
		},
		AllowedMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE",
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
		},
	}))

	r.MethodNotAllowed(handler.MethodNotAllowed)
	r.NotFound(handler.NotFound)

	// Rotes
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("Iagher!!!!!!!!!!!!!!!!!!!!!!!!!!!"))
	})

	registerAPIRoutes(r, handler)

	return r
}

func registerAPIRoutes(r chi.Router, handler *handlers.Handler) {
	registerProductRoutes(r, "/api/products", handler)
	registerPresencesRoutes(r, "/api/presences", handler)
	// TODO: Phase 2 security:
	// validate family invitation token and store authorized family in cookie/session.
}

func registerProductRoutes(r chi.Router, path string, handler *handlers.Handler) {
	r.Route(path, func(r chi.Router) {
		r.Post("/", handler.CreateProduct)
		r.Get("/", handler.GetProducts)
		r.Get("/{id}", handler.GetProduct)
		r.Put("/{id}", handler.UpdateProduct)
		r.Delete("/{id}", handler.DeleteProduct)
	})
}

func registerPresencesRoutes(r chi.Router, path string, handler *handlers.Handler) {
	r.Route(path, func(r chi.Router) {
		r.Post("/", handler.CreateConfirmationPresence)
		r.Get("/", handler.GetConfirmationPresences)
		r.Get("/{id}", handler.GetConfirmationPresence)
		r.Put("/{id}", handler.UpdateConfirmationPresence)
		r.Patch("/{id}/confirm", handler.ConfirmPresence)
		r.Patch("/{id}/cancel", handler.CancelPresence)
		r.Delete("/{id}", handler.DeleteConfirmationPresence)
	})
}
