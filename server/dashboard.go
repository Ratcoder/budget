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

	transactions, err := (*db).GetTransactionsDateRange(userId, "2023-12-01", "2023-12-31")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Categories     []database.Category
		TotalAvailable int
		TotalBudgeted  int
		Transactions   []database.Transaction
	}{
		Categories:     categories,
		TotalAvailable: totalAvailable,
		TotalBudgeted:  totalBudgeted,
		Transactions:   transactions,
	}
	view.Template.ExecuteTemplate(w, "dashboard.html", data)
}
