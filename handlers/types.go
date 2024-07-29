package handler

import (
	"chat-app/database"
)

type Message struct {
	Content string
	Sender  string
	Recipient string
	Date   string
	// helps with styling the message in the chat
	RecipientMessage bool
}

type LayoutData struct{
	FormData FormData
	PageData PageData
	Username string
	WebsocketKeys map[string]string
}

type PageData struct {
	User database.User
	SidebarUsers []database.SidebarUser
	Messages []Message
	TargetUser string
	FormData FormData
	WebsocketKey string
}


type UsersWithWebsocketKeys struct {
	Username1 string
	Username2 string
	WebsocketKey string
}

type FormData struct{
	Values map[string]string
	Error string
	TargetUser string
}