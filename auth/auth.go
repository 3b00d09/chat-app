package auth

import (
	"chat-app/database"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/xyproto/randomstring"
)

// TODO
// creating sessions

func CreateSession(userId string) http.Cookie {

	fmt.Println(userId)

	sessionId := randomstring.CookieFriendlyString(14)

	newSession := database.UserSession{
		ID:            sessionId,
		UserID:        "23153",
		ActiveExpires: time.Now().Add(1 * time.Minute).Unix(),
		IdleExpires:   0,
	}

	db, err := database.SetupDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	statement, err := db.Prepare("INSERT INTO user_session (id, user_id, active_expires, idle_expires) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	defer statement.Close()

	fmt.Println(newSession)

	_, err = statement.Exec(newSession.ID, newSession.UserID, newSession.ActiveExpires, newSession.IdleExpires)
	if err != nil {
		log.Fatal(err)
	}

	cookie := http.Cookie{
		Name:     "session_token",
		Value:    sessionId,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	return cookie

}

func AuthenticateSession(cookie string) bool {
	db, err := database.SetupDB()
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	statement, err := db.Prepare("SELECT * FROM user_session WHERE id = ?")

	if err != nil {
		log.Fatal(err)
	}

	row := statement.QueryRow(cookie)

	var sessionID, userID string
	var activeExpires, idleExpires int64

	err = row.Scan(&sessionID, &userID, &activeExpires, &idleExpires)
	if err != nil {
		log.Fatal(err)
	}

	return activeExpires < time.Now().Unix()

}
func UserExists(username string) bool {
	db, err := database.SetupDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	statement, err := db.Prepare("SELECT username FROM user WHERE username = ?")

	if err != nil {
		log.Fatal(err)
	}

	var name string
	err = statement.QueryRow(username).Scan(&name)

	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("No rows")
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

	hashedPassword := GeneratHashedPassword(user.Password)

	statement, err := db.Prepare("INSERT INTO user (id, username, password) VALUES (?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	defer statement.Close()
	uid, err := uuid.NewRandom()
	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec(uid.String(), user.Username, hashedPassword)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Successfully added %s", user.Username)

	CreateSession(uid.String())
}
