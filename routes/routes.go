package routes

import (
	handler "chat-app/handlers"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Get("/", handler.HandleIndexRoute)
	r.Get("/login", handler.HandleLoginRoute)
	r.Get("/register", handler.HandleRegisterRoute)

	return r
}
