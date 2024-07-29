package routes

import (
	handler "chat-app/handlers"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes() *chi.Mux {
	
	r := chi.NewRouter()

	// disable caching for all routes as it causes a lot of bugginess in login, logout, and chat routes
	r.Use(disableCache)

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if filepath.Ext(r.URL.Path) == ".css" {
				w.Header().Set("Content-Type", "text/css")
			}
			next.ServeHTTP(w, r)
		})
	})

	r.Get("/", handler.HandleIndexRoute)
	r.Get("/chat/{user}", handler.HandleChatRoute)
	r.Get("/login", handler.HandleLoginRoute)
	r.Post("/login", handler.HandleLoginSubmission)
	r.Get("/register", handler.HandleRegisterRoute)
	r.Post("/register", handler.HandleRegisterSubmission)
	r.Get("/logout", handler.HandleLogoutRoute)
	r.Get("/search", handler.HandleSearchRoute)
	r.Post("/send-message", handler.HandleSendMessage)
	r.Get("/ws/{key}", handler.HandleWebSocket)



	// Serve static files from the specified directory, has to be directory not file
	fs := http.FileServer(http.Dir("./assets"))
	r.Handle("/assets/*", http.StripPrefix("/assets/", fs))

	return r
}

func disableCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}
