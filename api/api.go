package api

import (
	"database/sql"
	"net/http"
	"strconv"
)

var db *sql.DB

func Start(port int, databasee *sql.DB) error {
	db = databasee

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/", fs)
	mux.Handle("/api/transactions", authMiddleware(http.HandlerFunc(transactions)))
	mux.Handle("/api/categories", authMiddleware(http.HandlerFunc(categories)))
	mux.Handle("/api/accounts", authMiddleware(http.HandlerFunc(accounts)))
	mux.HandleFunc("/api/register", register)
	mux.HandleFunc("/api/login", login)

	return http.ListenAndServe(":"+strconv.Itoa(port), mux)
}