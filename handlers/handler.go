package handler

import (
	"chat-app/auth"
	"chat-app/database"
	"fmt"
	"html/template"
	"log"
	"net/http"
)


type PageData struct {
	User database.User
}


func HandleIndexRoute(w http.ResponseWriter, r *http.Request) {

	templates := template.Must(template.ParseFiles("views/layout.html", "views/index.html"))
	var User database.User  = auth.AuthenticateRequest(w, r)

	data := PageData{
		User: User,
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
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &newCookie)
	http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	
}

func HandleLoginSubmission(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("email")
	user := database.UserCredentials{
		Username: username,
		Password: r.FormValue("password"),
	}
	if auth.UserExists(user.Username) {
		sessionCookie := auth.CreateSession(username)
		http.SetCookie(w, &sessionCookie)
		http.Redirect(w, r, "/", http.StatusSeeOther)

	}else{
		log.Fatal("invalid credentials")
	}
}

func HandleRegisterSubmission(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	username := r.FormValue("username")
	exists := auth.UserExists(username)
	if !exists {
		if r.FormValue("password1") == r.FormValue("password2") {
			user := database.UserCredentials{
				Username: username,
				Password: r.FormValue("password1"),
			}
			var cookie http.Cookie = auth.CreateUser(user)
			http.SetCookie(w, &cookie)
		} else {
			fmt.Print("Passwords Dont Match")
		}
	} else {
		fmt.Print("User Exists")
	}

}
