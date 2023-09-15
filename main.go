package main

import (
	"chat-app/routes"
	"chat-app/server"
	"context"
	"fmt"
	"net/http"

	_ "github.com/libsql/libsql-client-go/libsql"
	_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
)

const port string = ":3000"

func main() {
	r := routes.SetupRoutes()
	ctx := context.Background()

	db, err := server.SetupDB()
	if err != nil {
		fmt.Print(err)
	}
	data, err := db.QueryContext(ctx, "select * from users")
	if err != nil {
		fmt.Print(err)
	}
	for data.Next() {
		var id int
		var email string
		err := data.Scan(&id, &email)
		if err != nil {
			fmt.Print("Error scanning")
		}
		fmt.Printf("ID: %d, Username: %s\n", id, email)
	}
	fmt.Printf("Server Running on http://localhost%s\n", port)
	http.ListenAndServe(port, r)
}
