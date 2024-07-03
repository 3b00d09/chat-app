package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/libsql/libsql-client-go/libsql"
	_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
)
var DB *sql.DB

// the init function runs before main automatically
func init() {
	
	if err := godotenv.Load(); err != nil {
		fmt.Print("Failed to load .env")
		return
	}

	var err error
	DB, err = SetupDB()
	if err != nil{
		// handle properly later
		log.Fatal("Connection to database failed")
	}

	RunSchema(DB)

}

func SetupDB() (*sql.DB, error) {
    dbKey := os.Getenv("DB_KEY")
    dbUrl := fmt.Sprintf("libsql://chat-app-3b00d09.turso.io?authToken=%s", dbKey)
    return sql.Open("libsql", dbUrl)
}
