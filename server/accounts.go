package server

import (
	"budget/database"
	"budget/view"
	"net/http"
)

func accounts(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("user").(int)
	accounts, err := (*db).GetAccounts(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Accounts []database.Account
	}{
		Accounts: accounts,
	}
	view.Template.ExecuteTemplate(w, "accounts.html", data)
}
