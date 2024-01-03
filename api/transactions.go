package api

import (
	"net/http"
	"budget/database"
	"encoding/json"
	"io"
)

func transactions(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("user").(int)
	switch r.Method {
	case "GET":
		// Get all transactions
		transactions, err := (*db).GetTransactions(userId)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
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
		// Read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		// Convert to Transaction
		var transaction database.Transaction
		err = json.Unmarshal(body, &transaction)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		transaction.UserId = userId

		// Create transaction
		err = (*db).CreateTransaction(transaction)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write response
		w.WriteHeader(http.StatusCreated)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}