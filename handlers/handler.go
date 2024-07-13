package handler

import (
	"chat-app/auth"
	"chat-app/database"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

type Connection struct {
    ws   *websocket.Conn
    user string
}

var (
    connections = make(map[string]*Connection)
	// a mutex is mutual exclusion lock. It is used to synchronize access to shared resources so we dont get race conditions and locks a resource while its being used.
    connMutex   sync.Mutex
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

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	user := chi.URLParam(r, "user")
	// Upgrade the connection to a websocket connection
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}

	// Lock the resource and store the connection
    connMutex.Lock()
    connections[user] = &Connection{ws: conn, user: user}
    connMutex.Unlock()

    // Keep the connection alive
    for {
        _, _, err := conn.Read(context.Background())
        if err != nil {
			// If there is an error, break the loop which will make the func implicitly return, running the defer function
            break
        }
    }

	// Defer the removal of the connection when the function returns
    defer func() {
        connMutex.Lock()
        delete(connections, user)
        connMutex.Unlock()
        conn.Close(websocket.StatusNormalClosure, "Connection closed")
    }()

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

	// fire off the func that broadcasts the message to the recipient 
	broadcastMessage(User.Username, recepientUser.Username, message)

}

func broadcastMessage(sender, recipient, message string) {
    connMutex.Lock()

	// unlock connection when broadcast fun is done
    defer connMutex.Unlock()

    messageData := map[string]interface{}{
        "sender":  sender,
		"recepient": recipient,
        "message":   message,
        "timestamp": time.Now(),
    }

	// convert to JSON
    jsonMessage, _ := json.Marshal(messageData)

    // Send message
    if conn, ok := connections[sender]; ok {
        conn.ws.Write(context.Background(), websocket.MessageText, jsonMessage)
    }
}