package api

import (
	"net/http"
	"encoding/json"
	"budget/database"
)

type Account struct {
	Id            int    `json:"id,omitempty"`
	Name          string `json:"name"`
	Balance       int    `json:"balance"`
}

func accounts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getAccounts(w, r)
	case "POST":
		createAccount(w, r)
	case "PATCH":
		updateAccount(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getAccounts(w http.ResponseWriter, r *http.Request) {
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

func createAccount(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("user").(int)
	var apiAccount Account
	err := json.NewDecoder(r.Body).Decode(&apiAccount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	account := database.Account{
		Name:    apiAccount.Name,
		Balance: apiAccount.Balance,
		UserId:  userId,
	}

	err = (*db).CreateAccount(account)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func updateAccount(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("user").(int)
	var apiAccount Account
	err := json.NewDecoder(r.Body).Decode(&apiAccount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	account := database.Account{
		Id:      apiAccount.Id,
		Name:    apiAccount.Name,
		Balance: apiAccount.Balance,
		UserId:  userId,
	}

	err = (*db).UpdateAccount(userId, account)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}