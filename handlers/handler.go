package handler

import (
	"html/template"
	"net/http"
)

type User struct {
	LoggedIn bool
	Username string
}

func HandleIndexRoute(w http.ResponseWriter, r *http.Request) {
	templates := template.Must(template.ParseFiles("views/layout.html", "views/index.html"))
	user := User{
		LoggedIn: true,
		Username: "annon",
	}
	err := templates.ExecuteTemplate(w, "layout.html", user)
	if err != nil {
		http.Error(w, "Failed to parse template"+err.Error(), http.StatusInternalServerError)
		return
	}
}
