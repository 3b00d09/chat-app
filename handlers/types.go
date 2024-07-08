package handler

import (
	"chat-app/database"
)

type Message struct {
	Content string
	Sender  string
	Date   string
	// helps with styling the message in the chat
	RecipientMessage bool
}

type LayoutData struct{
	FormData FormData
	PageData PageData
	Username string
}

type PageData struct {
	User database.User
	Users []string
	Messages []Message
	TargetUser string
	FormData FormData
}

type FormData struct{
	Values map[string]string
	Error string
	TargetUser string
}