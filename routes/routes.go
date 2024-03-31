package routes

import (
	handler "chat-app/handlers"
	server "chat-app/server"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", handler.HandleIndexRoute)
	r.With(server.MiddlewareTest).Get("/login", handler.HandleLoginRoute)
	r.Post("/login", handler.HandleLoginSubmission)
	r.Get("/register", handler.HandleRegisterRoute)
	r.Post("/register", handler.HandleRegisterSubmission)
	r.Get("/logout", handler.HandleLogoutRoute)

	return r
}
