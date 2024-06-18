package api

import (
	"encoding/json"
	"io"
	"net/http"
)

type Transaction struct {
	Id            int    `json:"id,omitempty"`
	Year          int    `json:"year"`
	Month         int    `json:"month"`
	Day           int    `json:"day"`
	Description   string `json:"description"`
	Amount        int    `json:"amount"`
	AccountId     int    `json:"account_id"`
	CategoryId    int    `json:"category_id"`
	// IsTransfer    bool   `json:"is_transfer"`
}

func transactions(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("user").(int)
	switch r.Method {
	case "GET":
		// Get all transactions
		rows, err := db.Query("SELECT transaction_id, year, month, day, description, amount, transactions.account_id, category_id FROM transactions JOIN accounts ON accounts.account_id = transactions.account_id WHERE user_id = ?", userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		transactions := make([]Transaction, 0)
		for rows.Next() {
			var transaction Transaction
			err := rows.Scan(&transaction.Id, &transaction.Year, &transaction.Month, &transaction.Day, &transaction.Description, &transaction.Amount, &transaction.AccountId, &transaction.CategoryId)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			transactions = append(transactions, transaction)
		}

		// Convert to JSON
		jsonTransactions, err := json.Marshal(transactions)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		// Write response
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonTransactions)
	case "POST":
		// Create a new transaction
		var transaction Transaction
		err := json.NewDecoder(r.Body).Decode(&transaction)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		stmt, err := db.Prepare("INSERT INTO transactions(year, month, day, description, amount, account_id, category_id, user_id) VALUES(?, ?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = stmt.Exec(transaction.Year, transaction.Month, transaction.Day, transaction.Description, transaction.Amount, transaction.AccountId, transaction.CategoryId, userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Write response
		w.WriteHeader(http.StatusCreated)
	case "PATCH":
		// Update a transaction
		// Read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		// Parse body
		var transaction Transaction
		err = json.Unmarshal(body, &transaction)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		// Update transaction
		stmt, err := db.Prepare("UPDATE transactions SET year = ?, month = ?, day = ?, description = ?, amount = ?, account_id = ?, category_id = ? WHERE transaction_id = ?")
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		_, err = stmt.Exec(transaction.Year, transaction.Month, transaction.Day, transaction.Description, transaction.Amount, transaction.AccountId, transaction.CategoryId, transaction.Id)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		// Write response
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}