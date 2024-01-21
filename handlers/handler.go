package handler

import (
	"chat-app/auth"
	"chat-app/database"
	"fmt"
	"html/template"
	"log"
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
	r.ParseForm()
	username := r.FormValue("email")
	user := database.User{
		Username: username,
		Password: r.FormValue("password"),
	}
	if auth.UserExists(user.Username) {
		browserCookie, err := r.Cookie("session_token")

		if err != nil {
			if err == http.ErrNoCookie {
				sessionCookie := auth.CreateSession(username)
				fmt.Println(sessionCookie)
				http.SetCookie(w, &sessionCookie)
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			} else {
				log.Fatal("something went wrong")
			}
		}
		isValidSession := auth.AuthenticateSession(browserCookie.Value)

		if !isValidSession {
			fmt.Println("Session not valid")
		}

		// sessionCookie := auth.CreateSession(username)
		// http.SetCookie(w, &sessionCookie)
	}
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
