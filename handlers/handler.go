package handler

import (
	"chat-app/auth"
	"chat-app/database"
	"fmt"
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

func HandleRegisterRoute(w http.ResponseWriter, r *http.Request) {
	templates := template.Must(template.ParseFiles("views/layout.html", "views/register.html"))
	err := templates.ExecuteTemplate(w, "layout.html", nil)

	if err != nil {
		http.Error(w, "Failed to parse template"+err.Error(), http.StatusInternalServerError)
		return
	}
}

func HandleLoginRoute(w http.ResponseWriter, r *http.Request) {
	templates := template.Must(template.ParseFiles("views/layout.html", "views/login.html"))
	err := templates.ExecuteTemplate(w, "layout.html", nil)

	if err != nil {
		http.Error(w, "Failed to parse template"+err.Error(), http.StatusInternalServerError)
		return
	}
}

func HandleLoginSubmission(w http.ResponseWriter, r *http.Request) {
	fmt.Println("we are in login")
	r.ParseForm()
	username := r.FormValue("email")
	user := database.User{
		Username: username,
		Password: r.FormValue("password"),
	}
	auth.UserExists(user.Username)
}

func HandleRegisterSubmission(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	username := r.FormValue("username")
	exists := auth.UserExists(username)
	if !exists {
		if r.FormValue("password1") == r.FormValue("password2") {
			user := database.User{
				Username: username,
				Password: r.FormValue("password1"),
			}
			auth.CreateUser(user)
		} else {
			fmt.Print("Passwords Dont Match")
		}
	} else {
		fmt.Print("User Exists")
	}

}
