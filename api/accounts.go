package api

import (
	"net/http"
	"encoding/json"
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

	rows, err := db.Query("SELECT account_id, name FROM accounts WHERE user_id = ?", userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	accounts := make([]Account, 0)
	for rows.Next() {
		var account Account
		err = rows.Scan(&account.Id, &account.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err := db.QueryRow("SELECT SUM(amount) FROM transactions WHERE account_id = ?", account.Id).Scan(&account.Balance)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		accounts = append(accounts, account)
	}

	jsonAccounts, err := json.Marshal(accounts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonAccounts)
}

func createAccount(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("user").(int)
	var account Account
	err := json.NewDecoder(r.Body).Decode(&account)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stmt, err := db.Prepare("INSERT INTO accounts (user_id, name) VALUES (?, ?)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = stmt.Exec(userId, account.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func updateAccount(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("user").(int)
	var account Account
	err := json.NewDecoder(r.Body).Decode(&account)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stmt, err := db.Prepare("UPDATE accounts SET name = ? WHERE user_id = ? AND account_id = ?")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = stmt.Exec(account.Name, userId, account.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}