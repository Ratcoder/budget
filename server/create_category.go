package server

import (
	"budget/database"
	"budget/view"
	"net/http"
	"strconv"
)

func create_category(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		name := r.FormValue("name")
		available, err := strconv.Atoi(r.FormValue("available"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		budget, err := strconv.Atoi(r.FormValue("budgeted"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userId := r.Context().Value("user").(int)

		category := database.Category{
			Name:      name,
			Available: available,
			Budgeted:  budget,
			UserId:    userId,
		}
		err = (*db).CreateCategory(category)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/dashboard.html", http.StatusSeeOther)
		return
	}
	view.Template.ExecuteTemplate(w, "login.html", "")
}
