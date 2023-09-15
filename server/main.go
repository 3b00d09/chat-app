package main

import (
	"chat-app/routes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/libsql/libsql-client-go/libsql"
	_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
)

const port string = ":3000"

// the init function runs before main automatically
func init() {
	if err := godotenv.Load(); err != nil {
		fmt.Print("Failed to load .env")
	}
}

func main() {
	r := routes.SetupRoutes()
	ctx := context.Background()
	dbKey := os.Getenv("DB_KEY")
	dbUrl := fmt.Sprintf("libsql://chat-app-3b00d09.turso.io?authToken=%s", dbKey)
	db, err := sql.Open("libsql", dbUrl)
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
