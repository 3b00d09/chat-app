package routes

import (
	handler "chat-app/handlers"
	server "chat-app/server"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes() *chi.Mux {
	
	r := chi.NewRouter()

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if filepath.Ext(r.URL.Path) == ".css" {
				w.Header().Set("Content-Type", "text/css")
			}
			next.ServeHTTP(w, r)
		})
	})

	r.Get("/", handler.HandleIndexRoute)
	r.With(server.MiddlewareTest).Get("/login", handler.HandleLoginRoute)
	r.Post("/login", handler.HandleLoginSubmission)
	r.Get("/register", handler.HandleRegisterRoute)
	r.Post("/register", handler.HandleRegisterSubmission)
	r.Get("/logout", handler.HandleLogoutRoute)



	// Serve static files from the specified directory, has to be directory not file
	fs := http.FileServer(http.Dir("./assets"))
	r.Handle("/assets/*", http.StripPrefix("/assets/", fs))

	return r
}
