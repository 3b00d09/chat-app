package main

import (
	"chat-app/routes"
	"fmt"
	"net/http"
)

const port string = ":3000"

func main() {
	r := routes.SetupRoutes()
	fmt.Printf("Server Running on http://localhost%s\n", port)
	http.ListenAndServe(port, r)
}
