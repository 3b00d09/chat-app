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



func AuthenticateSession(cookie string) database.User {

	statement, err := database.DB.Prepare("SELECT * FROM user_session WHERE id = ?")

	if err != nil {
		log.Fatal(err)
	}
	defer statement.Close()

	row := statement.QueryRow(cookie)

	var sessionID, userID string
	var activeExpires, idleExpires int64

	err = row.Scan(&sessionID, &userID, &activeExpires, &idleExpires)
	if err != nil {
		return database.User{}
	}

	 if(activeExpires < time.Now().Unix()){
		return database.User{}
	 }

	 statement, err = database.DB.Prepare("SELECT id, username FROM user WHERE id = ?")

	 if err != nil {	
		 log.Fatal(err)
	 }
	 defer statement.Close()

	 row = statement.QueryRow(userID)
	 User := database.User{}
	 err = row.Scan(&User.ID, &User.Username)

	 if err != nil {
		 return database.User{}
	 }

	 return User	

}

func UserExists(username string) bool {

	statement, err := database.DB.Prepare("SELECT username FROM user WHERE username = ?")

	if err != nil {
		log.Fatal(err)
	}

	defer statement.Close()


	var name string
	err = statement.QueryRow(username).Scan(&name)

	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("User doesnt exist")
			return false
		}
		log.Fatal(err)
	}

	return true

}

func CreateUser(user database.UserCredentials) http.Cookie {

	hashedPassword := GeneratHashedPassword(user.Password)

	statement, err := database.DB.Prepare("INSERT INTO user (id, username, password) VALUES (?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	defer statement.Close()
	uid, err := uuid.NewRandom()
	if err != nil {
		log.Fatal(err)
	}

	rowId := uid.String()
	_, err = statement.Exec(rowId, user.Username, hashedPassword)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Successfully added %s", user.Username)

	var cookie http.Cookie = CreateSession(user.Username)

	return cookie
}

func CreateSession(username string) http.Cookie {

	query, err := database.DB.Prepare("SELECT id FROM user WHERE username = ?")

	if err != nil{
		log.Fatal(err)
	}

	var userId string

	err = query.QueryRow(username).Scan(&userId)

    if err != nil {
        if err == sql.ErrNoRows {
            fmt.Println("User not found")
        } else {
            log.Fatal(err)
        }
    }

	sessionId := randomstring.CookieFriendlyString(14)

	newSession := database.UserSession{
		ID:            sessionId,
		UserID:        userId,
		ActiveExpires: time.Now().Add(3600 * time.Minute).Unix(),
		IdleExpires:   0,
	}
	
	statement, err := database.DB.Prepare("INSERT INTO user_session (id, user_id, active_expires, idle_expires) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	defer statement.Close()

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

func ClearSession(token string){
	statement, err := database.DB.Prepare("DELETE FROM user_session WHERE id = ?")
	
	if err != nil{
		log.Fatal("Error deleting session")
	}
	defer statement.Close()

	statement.Exec(token)


}

func AuthenticateRequest(w http.ResponseWriter, r *http.Request) database.User {

	cookie, err := r.Cookie("session_token")

	if err != nil || cookie == nil{
		return database.User{}
	}

	user := AuthenticateSession(cookie.Value) 

	return user
}