package handler

import (
	"chat-app/auth"
	"chat-app/database"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Message struct {
	Content string
	Sender  string
	Date   string
}

type PageData struct {
	User database.User
	Users []string
	Messages []Message
	Chat bool
	TargetUser database.User
}

var messages []Message = []Message{}

func HandleIndexRoute(w http.ResponseWriter, r *http.Request) {

	var User database.User  = auth.AuthenticateRequest(w, r)

	if(User.ID == ""){
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}

	results := database.FetchSidebarUsers(User.Username)

	// still have to pass in chat.html because it will error out even if i have an if statement wrapping it in the template
	templates := template.Must(template.ParseFiles("views/layout.html", "views/index.html", "views/chat.html"))

	data := PageData{
		User: User,
		Users: results,
		Messages: messages,
		Chat: false,
	}

	err := templates.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		http.Error(w, "Failed to parse template"+err.Error(), http.StatusInternalServerError)
		return
	}
}

func HandleRegisterRoute(w http.ResponseWriter, r *http.Request) {
	templates := template.Must(template.ParseFiles("views/layout.html", "views/register.html"))
	var User database.User  = auth.AuthenticateRequest(w, r)

	if(User.ID != ""){
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}

	err := templates.ExecuteTemplate(w, "layout.html", nil)

	if err != nil {
		http.Error(w, "Failed to parse template"+err.Error(), http.StatusInternalServerError)
		return
	}
}

func HandleLoginRoute(w http.ResponseWriter, r *http.Request) {
	templates := template.Must(template.ParseFiles("views/layout.html", "views/login.html"))

	var User database.User  = auth.AuthenticateRequest(w, r)

	if(User.ID != ""){
		fmt.Print("User is already logged in")
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}

	err := templates.ExecuteTemplate(w, "layout.html", nil)
	if err != nil {
		http.Error(w, "Failed to parse template"+err.Error(), http.StatusInternalServerError)
		return
	}
}

func HandleLogoutRoute(w http.ResponseWriter, r *http.Request) {

   // Set Cache-Control headers to prevent caching from keeping the cookie alive :)))
    w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
    w.Header().Set("Pragma", "no-cache")
    w.Header().Set("Expires", "0")

	cookie, err := r.Cookie("session_token")

	if err != nil || cookie == nil{
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}

	auth.ClearSession(cookie.Value)

	newCookie := http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires: time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &newCookie)
	http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	
}

func HandleLoginSubmission(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	user := database.UserCredentials{
		Username: username,
		Password: r.FormValue("password"),
	}
	if auth.UserExists(user) {
		sessionCookie := auth.CreateSession(username)
		http.SetCookie(w, &sessionCookie)
		http.Redirect(w, r, "/", http.StatusSeeOther)

	}else{
		PageData := struct {
			// layout.html expects a User struct even if empty
			User database.User
			ErrorMessage string
			PreviousUsername string
		}{
			ErrorMessage: "Invalid Credentials",
			PreviousUsername: username,
		}
		templates := template.Must(template.ParseFiles("views/layout.html", "views/login.html"))
		err := templates.ExecuteTemplate(w, "layout.html", PageData)
		if err != nil {
			http.Error(w, "Failed to parse template"+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func HandleRegisterSubmission(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	username := r.FormValue("username")
	isUnique := auth.IsUniqueUsername(username)
	if isUnique {
		if r.FormValue("password1") == r.FormValue("password2") {
			user := database.UserCredentials{
				Username: username,
				Password: r.FormValue("password1"),
			}
			var cookie http.Cookie = auth.CreateUser(user)
			http.SetCookie(w, &cookie)
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
		} else {
			template := template.Must(template.ParseFiles("views/layout.html", "views/register.html"))
			PageData := struct {
				// layout.html expects a User struct even if empty
				User database.User
				ErrorMessage string
				PreviousUsername string
			}{
				ErrorMessage: "Passwords do not match",
				PreviousUsername: username,
			}
			err := template.ExecuteTemplate(w, "layout.html", PageData)
			if err != nil {
				http.Error(w, "Failed to parse template"+err.Error(), http.StatusInternalServerError)
				return
			}
		}
	} else {
		template := template.Must(template.ParseFiles("views/layout.html", "views/register.html"))
		PageData := struct {
			// layout.html expects a User struct even if empty
			User database.User
			ErrorMessage string
			PreviousUsername string
		}{
			ErrorMessage: "Username already exists",
		}
		err := template.ExecuteTemplate(w, "layout.html", PageData)
		if err != nil {
			http.Error(w, "Failed to parse template"+err.Error(), http.StatusInternalServerError)
			return
		}
	}

}

func HandleSearchRoute(w http.ResponseWriter, r *http.Request) {

	queryParam := r.URL.Query().Get("q")
	queryParam = strings.Trim(queryParam, " ")
	if(queryParam == ""){
		return
	}

	statement, err := database.DB.Prepare("SELECT username FROM user WHERE username LIKE ?")

	if err != nil {
		log.Fatal(err)
	}
	defer statement.Close()

	rows, err := statement.Query("%" + queryParam + "%")
	if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

	var results []string
    for rows.Next() {
        var username string
        err := rows.Scan(&username)
        if err != nil {
            log.Fatal(err)
        }
        results = append(results, username)
    }

	var html string

	if len(results) == 0 {
		html = "<div>No Results Found</div>"
	}else{
		for _, username := range results {
			html += fmt.Sprintf("<div>%s</div>", username)
		}
	}

	w.Write([]byte(html))
}

func HandleChatRoute(w http.ResponseWriter, r *http.Request) {
	targetUser := chi.URLParam(r, "user")

	templates := template.Must(template.ParseFiles("views/layout.html", "views/index.html", "views/chat.html"))
	var User database.User  = auth.AuthenticateRequest(w, r)

	if(User.ID == ""){
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}

	results := database.FetchSidebarUsers(User.Username)
	statement, err := database.DB.Prepare("SELECT id, username FROM user WHERE username = ?")

	if err != nil {
		log.Fatal(err)
	}
	defer statement.Close()

	rows := statement.QueryRow(targetUser)

	var recepientUser database.User
	err = rows.Scan(&recepientUser.ID, &recepientUser.Username)

	if err != nil {
		recepientUser.Username = ""
	}

	statement, err = database.DB.Prepare("SELECT user.username, messages.message, messages.created_at FROM messages LEFT JOIN user ON messages.user_id = user.id WHERE user.id IN (?, ?) ORDER BY messages.created_at ASC")

	if err != nil {
		log.Fatal(err)
	}

	rows2, err := statement.Query(User.ID, recepientUser.ID)

	if err != nil {
		log.Fatal(err)
	}

	for rows2.Next(){
		var message Message
		err := rows2.Scan(&message.Sender, &message.Content, &message.Date)
		if err != nil {
			fmt.Println(err)
		}
		messages = append(messages, message)
	}

	

	data := PageData{
		User: User,
		Users: results,
		Messages: messages,
		Chat: true,
		TargetUser: recepientUser,
	}

	err = templates.ExecuteTemplate(w, "layout.html", data)
	
	if err != nil {
		http.Error(w, "Failed to parse template"+err.Error(), http.StatusInternalServerError)
		return
	}

}

func HandleSendMessage(w http.ResponseWriter, r *http.Request){
	fmt.Println("Sending message")

	var User database.User  = auth.AuthenticateRequest(w, r)

	if(User.ID == ""){
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}

	r.ParseForm()
	message := r.FormValue("chat-message")
	if message == "" {
		return
	}

	targetUser := r.FormValue("target-user")

	if targetUser == "" {
		return
	}

	statement, err := database.DB.Prepare("SELECT id, username FROM user WHERE username = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer statement.Close()

	rows := statement.QueryRow(targetUser)

	var recepientUser database.User
	err = rows.Scan(&recepientUser.ID, &recepientUser.Username)

	if err != nil {
		fmt.Println("User not found")
	}

	statement, err = database.DB.Prepare("INSERT INTO messages (id, user_id, message, created_at) VALUES (?, ?, ?, ?)")

	if err != nil {
		log.Fatal(err)
	}


	newRowId := uuid.New().String()
	newRowId = newRowId[:8]
	_, err = statement.Exec(newRowId, User.ID, message, time.Now().Unix())

	if err != nil {
		fmt.Println("Error inserting message")
	}

	htmlResponse := fmt.Sprintf( `
	<div class="flex gap-4 items-center justify-end">
		<p class="bg-primary p-2 rounded w-1/2">%s</p>
		<p>%s</p>
	</div>`,  message, User.Username)


	w.Write([]byte(htmlResponse))

}