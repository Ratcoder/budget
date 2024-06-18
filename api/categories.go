package api

import (
	"net/http"
	"encoding/json"
)

type Category struct {
	Id            int                 `json:"id,omitempty"`
	Name          string              `json:"name"`
	Available     int                 `json:"available"`
	Assigned      int                 `json:"assigned"`
	// BudgetType    database.BudgetType `json:"budget_type"`
	// BudgetAmount  int	              `json:"budget_amount"`
}

func categories(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		createCategory(w, r)
		return
	} else if r.Method == "PATCH" {
		updateCategory(w, r)
		return
	} else if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userId := r.Context().Value("user").(int)
	rows, err := db.Query("SELECT category_id, name FROM categories WHERE user_id = ?", userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	categories := make([]Category, 0)
	for rows.Next() {
		var category Category
		err := rows.Scan(&category.Id, &category.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var transactionsAmount int
		var assignAmount int
		err = db.QueryRow("SELECT SUM(amount) FROM transactions WHERE category_id = ?", category.Id).Scan(&transactionsAmount)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = db.QueryRow("SELECT SUM(amount) FROM assign WHERE category_id = ?", category.Id).Scan(&assignAmount)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		category.Available = assignAmount + transactionsAmount
	}

	jsonCategories, err := json.Marshal(categories)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonCategories)
}

func createCategory(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("user").(int)
	var category Category
	err := json.NewDecoder(r.Body).Decode(&category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stmt, err := db.Prepare("INSERT INTO categories (name, user_id) VALUES (?, ?)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = stmt.Exec(category.Name, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func updateCategory(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("user").(int)
	var category Category
	err := json.NewDecoder(r.Body).Decode(&category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stmt, err := db.Prepare("UPDATE categories SET name = ? WHERE user_id = ? AND category_id = ?")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = stmt.Exec(category.Name, userId, category.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}