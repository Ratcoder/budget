package server

import (
	"budget/database"
	"net/http"
	"strconv"
)

var db *database.Database

func Start(port int, databasee *database.Database) error {
	db = databasee

	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("static")))
	mux.HandleFunc("/login.html", login)
	mux.HandleFunc("/register.html", register)
	mux.Handle("/transactions.html", authMiddleware(http.HandlerFunc(transactions)))
	mux.Handle("/accounts.html", authMiddleware(http.HandlerFunc(accounts)))
	mux.Handle("/dashboard.html", authMiddleware(http.HandlerFunc(dashboard)))
	mux.Handle("/create_category", authMiddleware(http.HandlerFunc(create_category)))

	mux.Handle("/api/get_link_token", authMiddleware(http.HandlerFunc(get_link_token)))
	mux.Handle("/api/set_public_token", authMiddleware(http.HandlerFunc(set_public_token)))

	return http.ListenAndServe(":"+strconv.Itoa(port), mux)
}
