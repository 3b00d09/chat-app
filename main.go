package main

import (
	"chat-app/database"
	"chat-app/routes"

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

	db, err := database.SetupDB()
	if err != nil {
		fmt.Print(err)
	}
	database.RunSchema(db)

	data, err := db.QueryContext(ctx, "select * from user")
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
