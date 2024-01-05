package api

import (
	"budget/database"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type Transaction struct {
	Id            int    `json:"id"`
	Date          string `json:"date"`
	Description   string `json:"description"`
	Amount        int    `json:"amount"`
	Account       string `json:"account"`
	CategoryId    int    `json:"category_id"`
}

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

		// Convert to api transactions
		var apiTransactions []Transaction = make([]Transaction, len(transactions))
		for i, transaction := range transactions {
			apiTransaction := Transaction{
				Id:            transaction.Id,
				Date:          transaction.Date,
				Description:   transaction.Description,
				Amount:        transaction.Amount,
				Account:       transaction.Account,
				CategoryId:    transaction.CategoryId,
			}
			apiTransactions[i] = apiTransaction
		}

		// Convert to JSON
		jsonTransactions, err := json.Marshal(apiTransactions)
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
	case "PATCH":
		// Update a transaction
		// Read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		// Convert to Api Transaction
		var apiTransaction Transaction
		err = json.Unmarshal(body, &apiTransaction)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		log.Println(apiTransaction)

		// Convert to database Transaction
		transaction := database.Transaction{
			Id:            apiTransaction.Id,
			Date:          apiTransaction.Date,
			Description:   apiTransaction.Description,
			Amount:        apiTransaction.Amount,
			Account:       apiTransaction.Account,
			CategoryId:    apiTransaction.CategoryId,
			UserId:        userId,
		}

		// Update transaction
		err = (*db).UpdateTransaction(userId, transaction)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write response
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}