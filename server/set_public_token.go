package server

import (
	plaid "budget/plaid_connection"
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

	// Store access token in database
	userId := r.Context().Value("user").(int)
	user, err := (*db).GetUserById(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user.PlaidItem = accessToken
	if err := (*db).UpdateUser(userId, user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
