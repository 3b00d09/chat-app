package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const port string = ":3000"

func main() {
	r := chi.NewRouter()
	fmt.Printf("Server Running on http://localhost%s\n", port)
	r.Get("/", indexRoute)

	http.ListenAndServe(port, r)
	fmt.Println("Server Running")
}

func indexRoute(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World!"))
}
