package server

import (
	"net/http"
	"budget/view"
)

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		sessionString, err := loginUser(username, password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			view.Template.ExecuteTemplate(w, "login.html", "Username or password incorrect")
			return
		}
		
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    sessionString,
			Path:     "/",
			// Secure:   false,
			// HttpOnly: true,
			// SameSite: http.SameSiteStrictMode,
		})
		http.Redirect(w, r, "/transactions.html", http.StatusSeeOther)
		return
	}
	view.Template.ExecuteTemplate(w, "login.html", "")
}