package api

import (
	"encoding/json"
	"net/http"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Error   string `json:"error,omitempty"`
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sessionString, err := loginUser(req.Username, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := LoginResponse{Error: "Username or password incorrect"}
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Set session cookie
	cookie := http.Cookie{
		Name:     "session",
		Value:    sessionString,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	resp := LoginResponse{}
	json.NewEncoder(w).Encode(resp)
}
