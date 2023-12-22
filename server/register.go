package server

import (
	"net/http"
	"budget/view"
	"budget/database"
	"golang.org/x/crypto/bcrypt"
)

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Create a new user
		// Read body
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Check if user exists
		_, err := (*db).GetUserByName(username)
		if err == nil {
			w.WriteHeader(http.StatusInternalServerError)
			view.Template.ExecuteTemplate(w, "register.html", "User already exists")
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			view.Template.ExecuteTemplate(w, "register.html", "Internal server error")
			return
		}
		password = string(hashedPassword)

		// Create user
		user := database.User{
			Name:     username,
			Password: password,
		}
		err = (*db).CreateUser(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		http.Redirect(w, r, "/login.html", http.StatusSeeOther)
		return
	}
	view.Template.ExecuteTemplate(w, "register.html", "")
}