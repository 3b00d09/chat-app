package server

import (
	"chat-app/database"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// TODO
// check if user exists
// creating sessions
func UserExists(username string) bool {
	db, err := database.SetupDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	statement, err := db.Prepare("SELECT 1 FROM user WHERE username = ?")

	if err != nil {
		log.Fatal(err)
	}
	defer statement.Close()

	var exists int
	err = statement.QueryRow(username).Scan(&exists)

	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		log.Fatal(err)
	}
	return true

}

func CreateUser(user database.User) {
	db, err := database.SetupDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	statement, err := db.Prepare("INSERT INTO user (id, username, password) VALUES (?, ?, ?),")
	if err != nil {
		log.Fatal(err)
	}

	defer statement.Close()
	uid := 329480392548
	_, err = statement.Exec(uid, user.Username, user.Password)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Successfully added %s", user.Username)
}
