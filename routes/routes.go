package routes

import (
	handler "chat-app/handlers"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Get("/", handler.HandleIndexRoute)

	// Add more routes as needed

	return r
}
