package server

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/libsql/libsql-client-go/libsql"
	_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
)

// the init function runs before main automatically
func init() {
	if err := godotenv.Load(); err != nil {
		fmt.Print("Failed to load .env")
	}
}

func SetupDB() (*sql.DB, error) {
	dbKey := os.Getenv("DB_KEY")
	dbUrl := fmt.Sprintf("libsql://chat-app-3b00d09.turso.io?authToken=%s", dbKey)
	db, err := sql.Open("libsql", dbUrl)
	if err != nil {
		return nil, err
	}

	return db, nil
}
