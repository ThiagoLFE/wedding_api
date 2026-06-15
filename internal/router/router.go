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
		},
		AllowedMethods: []string{
			"GET", "POST", "PUT", "DELETE",
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
		w.Write([]byte("Hello World!"))
	})

	r.Route("/products", func(r chi.Router) {
		r.Post("/", handler.CreateProduct)
		r.Get("/", handler.GetProducts)
		r.Get("/{id}", handler.GetProduct)
		r.Put("/{id}", handler.UpdateProduct)
		r.Delete("/{id}", handler.DeleteProduct)
	})

	return r
}
