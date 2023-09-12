package handler

import "net/http"

func HandleIndexRoute(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello index"))
	return
}
