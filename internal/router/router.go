package router

import (
	"net/http"

	"wedding_api/internal/auth"
	"wedding_api/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(handler *handlers.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:5173",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.MethodNotAllowed(handler.MethodNotAllowed)
	r.NotFound(handler.NotFound)

	// The health endpoint is also protected so every application endpoint uses
	// the same authentication boundary. Authentication endpoints are registered
	// separately below.
	r.With(auth.Middleware(handler)).Get("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("Server is working good"))
	})
	// Keep the original product-list path as a protected compatibility alias.
	r.With(auth.Middleware(handler)).Get("/products", handler.GetProducts)

	r.Route("/api", func(r chi.Router) {
		// The only unauthenticated routes.
		r.Post("/auth/admin/login", handler.AdminLogin)
		r.Post("/auth/exchange", handler.ExchangeFamilyToken)

		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(handler))

			r.Post("/auth/logout", handler.Logout)
			r.Get("/auth/me", handler.Me)

			r.Route("/admin", func(r chi.Router) {
				r.Use(auth.RequireRoleMiddleware(auth.RoleAdmin))
				registerFamilyRoutes(r, handler)
			})

			registerProductRoutes(r, "/products", handler)
			registerPresenceRoutes(r, "/presences", handler)

			// A family user can access only its own family and its own presence list.
			r.Get("/family", handler.GetMyFamily)
		})
	})

	return r
}

func registerFamilyRoutes(r chi.Router, handler *handlers.Handler) {
	r.Route("/families", func(r chi.Router) {
		r.Post("/", handler.CreateFamily)
		r.Get("/", handler.GetFamilies)
		r.Get("/{id}", handler.GetFamily)
		r.Put("/{id}", handler.UpdateFamily)
		r.Delete("/{id}", handler.DeleteFamily)
		r.Post("/{id}/access-link", handler.CreateFamilyAccessLink)
		r.Post("/{id}/access-link/revoke", handler.RevokeFamilyAccessLink)
	})
}

func registerProductRoutes(r chi.Router, path string, handler *handlers.Handler) {
	r.Route(path, func(r chi.Router) {
		r.Get("/", handler.GetProducts)
		r.Get("/{id}", handler.GetProduct)

		r.Group(func(r chi.Router) {
			r.Use(auth.RequireRoleMiddleware(auth.RoleAdmin))
			r.Post("/", handler.CreateProduct)
			r.Put("/{id}", handler.UpdateProduct)
			r.Delete("/{id}", handler.DeleteProduct)
		})

		r.With(auth.RequireAnyRoleMiddleware(auth.RoleAdmin, auth.RoleFamily)).Patch("/{id}/reserve", handler.ReserveProduct)
	})
}

func registerPresenceRoutes(r chi.Router, path string, handler *handlers.Handler) {
	r.Route(path, func(r chi.Router) {
		r.Get("/", handler.GetConfirmationPresences)
		r.Get("/{id}", handler.GetConfirmationPresence)

		r.Group(func(r chi.Router) {
			r.Use(auth.RequireRoleMiddleware(auth.RoleAdmin))
			r.Post("/", handler.CreateConfirmationPresence)
			r.Put("/{id}", handler.UpdateConfirmationPresence)
			r.Delete("/{id}", handler.DeleteConfirmationPresence)
		})

		r.With(auth.RequireAnyRoleMiddleware(auth.RoleAdmin, auth.RoleFamily)).Patch("/{id}/confirm", handler.ConfirmPresence)
		r.With(auth.RequireAnyRoleMiddleware(auth.RoleAdmin, auth.RoleFamily)).Patch("/{id}/cancel", handler.CancelPresence)
	})
}
