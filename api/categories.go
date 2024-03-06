package api

import (
	"net/http"
	"encoding/json"
	"budget/database"
)

type Category struct {
	Id            int                 `json:"id,omitempty"`
	Name          string              `json:"name"`
	Available     int                 `json:"available"`
	Assigned      int                 `json:"assigned"`
	BudgetType    database.BudgetType `json:"budget_type"`
	BudgetAmount  int	              `json:"budget_amount"`
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
			Assigned:      category.Assigned,
			BudgetType:    category.BudgetType,
			BudgetAmount:  category.BudgetAmount,
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

func createCategory(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("user").(int)
	var apiCategory Category
	err := json.NewDecoder(r.Body).Decode(&apiCategory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	category := database.Category{
		Name:          apiCategory.Name,
		Available:     apiCategory.Available,
		Assigned:      apiCategory.Assigned,
		BudgetType:    apiCategory.BudgetType,
		BudgetAmount:  apiCategory.BudgetAmount,
		UserId:        userId,
	}

	err = (*db).CreateCategory(category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func updateCategory(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("user").(int)
	var apiCategory Category
	err := json.NewDecoder(r.Body).Decode(&apiCategory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	category := database.Category{
		Id:            apiCategory.Id,
		Name:          apiCategory.Name,
		Available:     apiCategory.Available,
		Assigned:      apiCategory.Assigned,
		BudgetType:    apiCategory.BudgetType,
		BudgetAmount:  apiCategory.BudgetAmount,
		UserId:        userId,
	}

	err = (*db).UpdateCategory(userId, category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}