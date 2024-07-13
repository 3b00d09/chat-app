package handler

import (
	"chat-app/auth"
	"chat-app/database"
	"html/template"
	"net/http"
	"time"
)


func HandleRegisterRoute(w http.ResponseWriter, r *http.Request) {
	templates := template.Must(template.ParseFiles("views/layout.html", "views/register.html"))
	var User database.User  = auth.AuthenticateRequest(w, r)

	if(User.ID != ""){
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}
	formData := FormData{
		Values: map[string]string{},
		Error: "",
	}

	LayoutData := LayoutData{
		FormData: formData,
		Username: "",
	}

	err := templates.ExecuteTemplate(w, "layout.html", LayoutData)

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

	formData := FormData{
		Values: map[string]string{},
		Error: "",
	}

	LayoutData := LayoutData{
		FormData: formData,
		Username: "",
	}

	err := templates.ExecuteTemplate(w, "layout.html", LayoutData)
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
		w.WriteHeader(http.StatusUnprocessableEntity)
		FormData := FormData{
				Values: map[string]string{
					"PreviousUsername": username,
				},
				Error: "Invalid username or password",
			}
		templates := template.Must(template.ParseFiles("views/login.html"))
		templates.ExecuteTemplate(w, "login-form", FormData)
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
			cookie, err := auth.CreateUser(user)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				FormData := FormData{
					Values: map[string]string{
						"PreviousUsername": username,
					},
					Error: "Something went wrong on the server. Please try again later.",
				}
				template := template.Must(template.ParseFiles("views/register.html"))
				template.ExecuteTemplate(w, "register-form", FormData)
			}
			http.SetCookie(w, &cookie)
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
		} else {
				w.WriteHeader(http.StatusUnprocessableEntity)
				FormData := FormData{
					Values: map[string]string{
						"PreviousUsername": username,
					},
					Error: "Passwords don't match.",
				}
				template := template.Must(template.ParseFiles("views/register.html"))
				template.ExecuteTemplate(w, "register-form", FormData)
		}
	} else {
		w.WriteHeader(http.StatusUnprocessableEntity)
		template := template.Must(template.ParseFiles("views/register.html"))
		FormData := FormData{
			Values: map[string]string{
				"PreviousUsername": username,
			},
			Error: "Username already exists",
		}
		template.ExecuteTemplate(w, "register-form", FormData)
	}

}