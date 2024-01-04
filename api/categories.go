package api

import (
	"net/http"
	"encoding/json"
)

type Category struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	Available     int    `json:"available"`
	Budgeted      int    `json:"budgeted"`
}

func categories(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userId := r.Context().Value("user").(int)
	categories, err := (*db).GetCategories(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var apiCategories []Category = make([]Category, len(categories))
	for i, category := range categories {
		apiCategories[i] = Category{
			Id:            category.Id,
			Name:          category.Name,
			Available:     category.Available,
			Budgeted:      category.Budgeted,
		}
	}

	jsonCategories, err := json.Marshal(apiCategories)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonCategories)
}