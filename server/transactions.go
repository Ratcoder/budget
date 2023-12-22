package server

import (
	"budget/database"
	"budget/view"
	"net/http"
)

func transactions(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("user").(int)
	transactions, err := (*db).GetTransactions(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Transactions []database.Transaction
	}{
		Transactions: transactions,
	}
	view.Template.ExecuteTemplate(w, "transactions.html", data)
}
