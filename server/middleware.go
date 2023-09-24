package server

import (
	"fmt"
	"net/http"
)

func MiddlewareTest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Logged: %s %s \n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
