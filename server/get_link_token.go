package server

import (
	plaid "budget/plaid_connection"
	"net/http"
)

func get_link_token(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userId := r.Context().Value("user").(int)
	linkToken, err := plaid.CreateLinkToken(userId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(linkToken))
}
