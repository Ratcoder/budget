package api

import (
	// "fmt"
	"budget/database"
	"net/http"
	"strconv"

	// "strings"
)

var db *database.Database

func Start(port int, databasee *database.Database) error {
	db = databasee

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/", fs)
	mux.Handle("/api/transactions", authMiddleware(http.HandlerFunc(transactions)))
	mux.HandleFunc("/api/register", register)
	mux.HandleFunc("/api/login", login)
	mux.Handle("/api/get_link_token", authMiddleware(http.HandlerFunc(get_link_token)))
	mux.Handle("/api/set_public_token", authMiddleware(http.HandlerFunc(set_public_token)))

	return http.ListenAndServe(":"+strconv.Itoa(port), mux)
}

// func cors(fs http.Handler) http.HandlerFunc {
//     return func(w http.ResponseWriter, r *http.Request) {
//         // do your cors stuff
//         // return if you do not want the FileServer handle a specific request

//         fs.ServeHTTP(w, r)
//     }
// }