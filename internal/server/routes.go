package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/wbrijesh/identity/internal/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)

	r.Get("/", s.HelloWorldHandler)
	r.Get("/health", s.healthHandler)

	// Admin routes
	r.Post("/admin/register", s.CreateAdminHandler)
	r.Post("/admin/login", s.LoginAdminHandler)

	// Application routes (protected by Admin auth middleware)
	r.Group(func(r chi.Router) {
		r.Use(middleware.AdminAuthMiddleware)

		r.Post("/applications", s.CreateApplicationHandler)
		r.Get("/applications", s.ListApplicationsHandler)
		r.Post("/applications/{applicationID}/refresh-token", s.GenerateRefreshTokenForApplicationHandler)
		r.Put("/applications/{applicationID}/refresh-token", s.UpdateRefreshTokenForApplicationHandler)
	})

	// User routes (protected by Access Token auth middleware)
	r.Group(func(r chi.Router) {
		r.Use(middleware.AcessTokenAuthMiddleware)

		r.Post("/users", s.CreateUserHandler)
		r.Post("/users/login", s.LoginUserHandler)
		r.Get("/applications/{applicationID}/users", s.ListUsersHandler)
	})

	return r
}
