package api

import (
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"database/sql"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Message string `json:"message"`
}

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Read body
	var req RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RegisterResponse{Message: "Invalid request"})
		return
	}

	// Check if user exists
	err = db.QueryRow("SELECT * FROM users WHERE name = ?", req.Username).Scan()
	if err != sql.ErrNoRows {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(RegisterResponse{Message: "User already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RegisterResponse{Message: "Internal server error"})
		return
	}
	password := string(hashedPassword)

	// Create user
	stmt, err := db.Prepare("INSERT INTO users(name, password) VALUES(?, ?)")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RegisterResponse{Message: "Failed to create user"})
		return
	}
	_, err = stmt.Exec(req.Username, password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RegisterResponse{Message: "Failed to create user"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(RegisterResponse{Message: "User created successfully"})
}
