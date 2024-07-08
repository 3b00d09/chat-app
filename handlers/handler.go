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


var messages []Message = []Message{}

func HandleIndexRoute(w http.ResponseWriter, r *http.Request) {

	var User database.User  = auth.AuthenticateRequest(w, r)

	if(User.ID == ""){
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}

	results := database.FetchSidebarUsers(User.Username)

	// still have to pass in chat.html because it will error out even if i have an if statement wrapping it in the template
	templates := template.Must(template.ParseFiles("views/layout.html", "views/index.html", "views/chat.html", "templates/form.html", "templates/message.html"))

	data := LayoutData{
		PageData: PageData{
			User: User,
			Users: results,
			Messages: messages,
			TargetUser: "",
		},
		Username: User.Username,
	}

	err := templates.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		http.Error(w, "Failed to parse template"+err.Error(), http.StatusInternalServerError)
		return
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

	templates := template.Must(template.ParseFiles("views/layout.html", "views/index.html", "views/chat.html", "templates/form.html", "templates/message.html"))
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
		if message.Sender == User.Username {
			message.RecipientMessage = false
		}else{
			message.RecipientMessage = true
		}
		messages = append(messages, message)
	}


	LayoutData := LayoutData{
		PageData: PageData{
			User: User,
			Users: results,
			Messages: messages,
			TargetUser: recepientUser.Username,
		},
		FormData: FormData{
				Values: map[string]string{},
				TargetUser: recepientUser.Username,
			},
		Username: User.Username,
	}

	err = templates.ExecuteTemplate(w, "layout.html", LayoutData)
	
	if err != nil {
		http.Error(w, "Failed to parse template"+err.Error(), http.StatusInternalServerError)
		return
	}

}

func HandleSendMessage(w http.ResponseWriter, r *http.Request){

	var User database.User  = auth.AuthenticateRequest(w, r)

	if(User.ID == ""){
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}

	r.ParseForm()
	formTemplate := template.Must(template.ParseFiles("templates/form.html"))
	
	formData := FormData{
		Values: map[string]string{},
		TargetUser: r.FormValue("target-user"),
		Error: "",
	}

	message := r.FormValue("chat-message")
	if message == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		formData.Error = "Message cannot be empty"
		formTemplate.Execute(w, formData)
		return
	}

	targetUser := r.FormValue("target-user")

	if targetUser == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		formData.Error = "Recipient not found, please refresh the page and try again."
		formTemplate.Execute(w, formData)
		return
	}

	statement, err := database.DB.Prepare("SELECT id, username FROM user WHERE username = ?")
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		formData.Error = "Something went wrong. Please try again later"
		formTemplate.Execute(w, formData)
		return
	}

	defer statement.Close()

	rows := statement.QueryRow(targetUser)

	var recepientUser database.User
	err = rows.Scan(&recepientUser.ID, &recepientUser.Username)

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		formData.Error = "Recipient not found, please refresh the page and try again."
		formTemplate.Execute(w, formData)
		return
	}

	statement, err = database.DB.Prepare("INSERT INTO messages (id, user_id, message, created_at) VALUES (?, ?, ?, ?)")

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		formData.Error = "Something went wrong. Please try again later."
		formTemplate.Execute(w, formData)
		return
	}


	newRowId := uuid.New().String()
	newRowId = newRowId[:8]
	timestamp := time.Now().Unix()
	_, err = statement.Exec(newRowId, User.ID, message, timestamp)

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		formData.Error = "Something went wrong. Please try again later."
		formTemplate.Execute(w, formData)
		return
	}

	messageTemplate := template.Must(template.ParseFiles("templates/message.html"))

	messageData := Message{
		Content: message,
		Sender: User.Username,
		Date: string(timestamp),
		RecipientMessage: false,
	}

	// only pulls out the "oob-message" block from the message.html template, and passes it messageData
	messageTemplate.ExecuteTemplate(w, "oob-message", messageData)
	formTemplate.Execute(w, formData)
	
} 