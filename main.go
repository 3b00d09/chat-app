package main

import (
	"chat-app/routes"
	"fmt"
	"net/http"

	_ "github.com/libsql/libsql-client-go/libsql"
	_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
)

const port string = ":3000"

func main() {
	r := routes.SetupRoutes()
	fmt.Printf("Server Running on http://localhost%s\n", port)
	http.ListenAndServe(port, r)
}
