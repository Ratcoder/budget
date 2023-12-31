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

	totalAvailable, totalBudgeted := 0, 0
	for _, category := range categories {
		totalAvailable += category.Available
		totalBudgeted += category.Budgeted
	}

	data := struct {
		Categories []database.Category
		TotalAvailable int
		TotalBudgeted int
	}{
		Categories: categories,
		TotalAvailable: totalAvailable,
		TotalBudgeted: totalBudgeted,
	}
	view.Template.ExecuteTemplate(w, "dashboard.html", data)
}
