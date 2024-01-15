package api

import (
	plaid "budget/plaid_connection"
	database "budget/database"
	"io"
	"net/http"
)

func set_public_token(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode public token from body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	publicToken := string(body)

	// Exchange public token for access token
	accessToken, err := plaid.ExchangePublicToken(publicToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Store access token in database
	userId := r.Context().Value("user").(int)
	item := database.PlaidItem{
		UserId: userId,
		AccessToken: accessToken,
		TransactionsCursor: "",
	}
	err = (*db).CreatePlaidItem(item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
