package api

import (
	"net/http"
	"encoding/json"
)

type Account struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	Balance       int    `json:"balance"`
}

func accounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userId := r.Context().Value("user").(int)
	accounts, err := (*db).GetAccounts(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var apiAccounts []Account = make([]Account, len(accounts))
	for i, account := range accounts {
		apiAccounts[i] = Account{
			Id:            account.Id,
			Name:          account.Name,
			Balance:       account.Balance,
		}
	}

	jsonAccounts, err := json.Marshal(apiAccounts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonAccounts)
}