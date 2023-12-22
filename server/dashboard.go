package server

import (
	"budget/database"
	"budget/view"
	"net/http"
)

func dashboard(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("user").(int)
	categories, err := (*db).GetCategories(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Categories []database.Category
	}{
		Categories: categories,
	}
	view.Template.ExecuteTemplate(w, "dashboard.html", data)
}
