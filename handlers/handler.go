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
	websocketKey string
}

type SidebarUser = database.SidebarUser

var (
	// using a slice of connections because the key by itself isnt enough to identify a connection causing an error where only the first connected user is stored
    connections = make(map[string][]*Connection)
	websocketsMap = make(map[string]string)
	// a mutex is mutual exclusion lock. It is used to synchronize access to shared resources so we dont get race conditions and locks a resource while its being used.
    connMutex   sync.Mutex
)


func HandleIndexRoute(w http.ResponseWriter, r *http.Request) {

	var User database.User  = auth.AuthenticateRequest(w, r)

	if(User.ID == ""){
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}

	SidebarUsers := database.FetchSidebarUsers(User.ID)

	// still have to pass in chat.html because it will error out even if i have an if statement wrapping it in the template
	templates := template.Must(template.ParseFiles("views/layout.html", "views/index.html", "views/chat.html", "templates/form.html", "templates/message.html"))

	data := LayoutData{
		PageData: PageData{
			User: User,
			SidebarUsers: SidebarUsers,
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
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	statement, err := database.DB.Prepare("SELECT id, username, websocket_key FROM user WHERE username = ?")

	if err != nil {
		log.Fatal(err)
	}
	defer statement.Close()

	rows := statement.QueryRow(targetUser)

	var recepientUser database.User
	err = rows.Scan(&recepientUser.ID, &recepientUser.Username, &recepientUser.WebsocketKey)

	if err != nil {
		recepientUser.Username = ""
	}

	// fetch conversation id automatically creates a new row for us if a conversation doesnt exist
	conversationId := database.FetchConversationId(User.ID, recepientUser.ID)

	SidebarUsers := database.FetchSidebarUsers(User.ID)

	// makes it much easier to parse it in js 
	websocketMapKey := fmt.Sprintf("%s,%s", User.Username, recepientUser.Username)

	// check if these two users already have a websocket connection
	exists := websocketsMap[websocketMapKey];
	if exists == "" {
		statement, err = database.DB.Prepare("SELECT websocket_key FROM user WHERE username = ?")
		if err != nil {
			log.Fatal(err)
		}
		var userWebsocketKey string
		rows = statement.QueryRow(User.Username)
		rows.Scan(&userWebsocketKey)

		websocketKey := database.GenerateCommonWebsocketKey(userWebsocketKey, recepientUser.WebsocketKey)
		websocketsMap[websocketMapKey] = websocketKey
	}

	statement, err = database.DB.Prepare("SELECT user.username, messages.message, messages.created_at FROM messages LEFT JOIN user ON messages.message_author = user.id LEFT JOIN conversations ON conversations.user1 = user.id OR conversations.user2 = user.id WHERE conversations.id = ? ORDER BY messages.created_at ASC")


	if err != nil {
		log.Fatal(err)
	}

	rows2, err := statement.Query(conversationId)

	if err != nil {
		log.Fatal(err)
	}

	var messages []Message = []Message{}
	for rows2.Next(){
		var message Message
		err := rows2.Scan(&message.Sender, &message.Content, &message.Date)
		if err != nil {
			fmt.Println(err)
		}
		message.RecipientMessage = message.Sender != User.Username

		messages = append(messages, message)
	}


	LayoutData := LayoutData{
		PageData: PageData{
			User: User,
			SidebarUsers: SidebarUsers,
			Messages: messages,
			TargetUser: recepientUser.Username,
			WebsocketKeys: websocketsMap,
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
	key := chi.URLParam(r, "key")

	// Upgrade the connection to a websocket connection
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}

	// Lock the resource and store the connection
    connMutex.Lock()
	connections[key] = append(connections[key], &Connection{ws: conn, websocketKey: key})
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
        delete(connections, key)
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

	statement, err := database.DB.Prepare("SELECT id, username, websocket_key FROM user WHERE username = ?")
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		formData.Error = "Something went wrong. Please try again later"
		formTemplate.Execute(w, formData)
		return
	}

	defer statement.Close()

	rows := statement.QueryRow(targetUser)

	var recepientUser database.User
	err = rows.Scan(&recepientUser.ID, &recepientUser.Username, &recepientUser.WebsocketKey)

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		formData.Error = "Recipient not found, please refresh the page and try again."
		formTemplate.Execute(w, formData)
		return
	}

	conversationId := database.FetchConversationId(User.ID, recepientUser.ID)

	statement, err = database.DB.Prepare("INSERT INTO messages (id, conversation_id, message_author, message, created_at) VALUES (?, ?, ?, ?, ?)")

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		formData.Error = "Something went wrong. Please try again later."
		formTemplate.Execute(w, formData)
		return
	}


	newRowId := uuid.New().String()
	newRowId = newRowId[:8]
	timestamp := time.Now().Unix()
	_, err = statement.Exec(newRowId, conversationId, User.ID, message, timestamp)

	if err != nil {
		fmt.Println(err)
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
	websocketKey := database.GenerateCommonWebsocketKey(User.WebsocketKey, recepientUser.WebsocketKey)
	broadcastMessage(User.Username, recepientUser.Username, message, websocketKey)

}

func broadcastMessage(sender, recepient, message, websocketKey string) {
    connMutex.Lock()

	// unlock connection when broadcast fun is done
    defer connMutex.Unlock()

    messageData := map[string]interface{}{
        "sender":  sender,
		"recepient": recepient,
        "message":   message,
        "timestamp": time.Now(),
    }

	// convert to JSON
    jsonMessage, _ := json.Marshal(messageData)

    // Send message
	for _, conn := range connections[websocketKey] {
        conn.ws.Write(context.Background(), websocket.MessageText, jsonMessage)
    }
}