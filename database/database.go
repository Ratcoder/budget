package database

import (
	"database/sql"
	"encoding/json"

	_ "github.com/mattn/go-sqlite3"
)

type Database interface {
	Init() error
	Close() error
	// Users
	CreateUser(u User) error
	GetUserByName(name string) (User, error)
	GetUserById(id int) (User, error)
	UpdateUser(id int, u User) error
	DeleteUser(id int) error
	// Transactions
	CreateTransaction(t Transaction) error
	GetTransactions(user int) ([]Transaction, error)
	UpdateTransaction(id int, t Transaction) error
	DeleteTransaction(id int) error
	// Categories
	CreateCategory(c Category) error
	GetCategories(user int) ([]Category, error)
	UpdateCategory(id int, name string) error
	DeleteCategory(id int) error
	// Budgets
	CreateBudget(b Budget) error
	GetBudgets(user int) ([]Budget, error)
	UpdateBudget(id int, b Budget) error
	DeleteBudget(id int) error
}

type Transaction struct {
	Id          int
	Date        string
	Description string
	Amount      int
	Account     string
	UserId      int
	PlaidCategory string
	CategoryId    int
}

type Category struct {
	Id        int
	Name      string
	UserId    int
	Available int
	Budgeted  int
}

type User struct {
	Id                      int
	Name                    string
	PlaidItem               string
	PlaidTransactionsCursor string
	Password                string
}

type Budget struct {
	Id        int
	UserId    int
	YearMonth string
	Items     []BudgetItem
}

type BudgetItem struct {
	CategoryId int
	Amount     int
}

type SqliteDB struct {
	driver *sql.DB
}

func CreateSqliteDB() *SqliteDB {
	return &SqliteDB{}
}

func (db *SqliteDB) Init() error {
	var err error
	db.driver, err = sql.Open("sqlite3", "./db")

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY,
		date TEXT,
		description TEXT,
		amount INT,
		account TEXT,
		user_id INT NOT NULL,
		plaid_category TEXT,
		category_id INT,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);
	-- DELETE FROM transactions;

	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY,
		name TEXT,
		user_id INT NOT NULL,
		available INT,
		budgeted INT,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);
	-- DELETE FROM categories;

	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		name TEXT,
		password TEXT,
		plaid_item TEXT,
		plaid_transactions_cursor TEXT
	);
	-- DELETE FROM users;

	CREATE TABLE IF NOT EXISTS budgets (
		id INTEGER PRIMARY KEY,
		user_id INT NOT NULL,
		year_month TEXT,
		items TEXT,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);
	-- PRAGMA foreign_keys = ON;
	`

	_, err = db.driver.Exec(sqlStmt)
	if err != nil {
		return err
	}

	return err
}

func (db *SqliteDB) Close() error {
	return db.driver.Close()
}

func (db *SqliteDB) CreateUser(u User) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO users(name, password, plaid_item, plaid_transactions_cursor) VALUES(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Name, u.Password, u.PlaidItem, u.PlaidTransactionsCursor)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *SqliteDB) GetUserByName(name string) (User, error) {
	var u User
	err := db.driver.QueryRow("SELECT * FROM users WHERE name = (?);", name).Scan(&u.Id, &u.Name, &u.Password, &u.PlaidItem, &u.PlaidTransactionsCursor)
	if err != nil {
		return u, err
	}

	return u, nil
}

func (db *SqliteDB) GetUserById(id int) (User, error) {
	var u User
	err := db.driver.QueryRow("SELECT * FROM users WHERE id = (?);", id).Scan(&u.Id, &u.Name, &u.Password, &u.PlaidItem, &u.PlaidTransactionsCursor)
	if err != nil {
		return u, err
	}

	return u, nil
}

func (db *SqliteDB) UpdateUser(id int, u User) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("UPDATE users SET name = (?), password = (?), plaid_item = (?), plaid_transactions_cursor = (?) WHERE id = (?);")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Name, u.Password, u.PlaidItem, u.PlaidTransactionsCursor, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *SqliteDB) DeleteUser(id int) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("DELETE FROM users WHERE id = (?);")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *SqliteDB) CreateTransaction(t Transaction) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO transactions(date, description, amount, account, user_id, plaid_category, category_id) VALUES(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(t.Date, t.Description, t.Amount, t.Account, t.UserId, t.PlaidCategory, t.CategoryId)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *SqliteDB) GetTransactions(user int) ([]Transaction, error) {
	rows, err := db.driver.Query("SELECT * FROM transactions WHERE user_id = (?);", user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction

	for rows.Next() {
		var t Transaction
		err := rows.Scan(&t.Id, &t.Date, &t.Description, &t.Amount, &t.Account, &t.UserId, &t.PlaidCategory, &t.CategoryId)
		if err != nil {
			return nil, err
		} else {
			transactions = append(transactions, t)
		}
	}

	return transactions, nil
}

func (db *SqliteDB) UpdateTransaction(id int, t Transaction) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("UPDATE transactions SET date = (?), description = (?), amount = (?), category_id = (?) WHERE id = (?);")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(t.Date, t.Description, t.Amount, t.CategoryId, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *SqliteDB) DeleteTransaction(id int) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("DELETE FROM transactions WHERE id = (?);")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *SqliteDB) CreateCategory(c Category) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO categories(name, user_id) VALUES(?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(c.Name, c.UserId)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *SqliteDB) GetCategories(user int) ([]Category, error) {
	rows, err := db.driver.Query("SELECT * FROM categories WHERE user_id = (?);", user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category

	for rows.Next() {
		var id int
		var name string
		err := rows.Scan(&id, &name)
		if err != nil {
			return nil, err
		} else {
			categories = append(categories, Category{Name: name})
		}
	}

	return categories, nil
}

func (db *SqliteDB) UpdateCategory(id int, name string) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("UPDATE categories SET name = (?) WHERE id = (?);")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *SqliteDB) DeleteCategory(id int) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("DELETE FROM categories WHERE id = (?);")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func serializeBudgetItems(items []BudgetItem) (string, error) {
	b, err := json.Marshal(items)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func deserializeBudgetItems(items string) ([]BudgetItem, error) {
	var budgetItems []BudgetItem
	err := json.Unmarshal([]byte(items), &budgetItems)
	if err != nil {
		return nil, err
	}
	return budgetItems, nil
}

func (db *SqliteDB) CreateBudget(b Budget) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	items, err := serializeBudgetItems(b.Items)
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO budgets(user_id, year_month, items) VALUES(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(b.UserId, b.YearMonth, items)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *SqliteDB) GetBudgets(user int) ([]Budget, error) {
	rows, err := db.driver.Query("SELECT * FROM budgets WHERE user_id = (?);", user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []Budget

	for rows.Next() {
		var b Budget
		var items string
		err := rows.Scan(&b.Id, &b.UserId, &b.YearMonth, &items)
		if err != nil {
			return nil, err
		} else {
			b.Items, err = deserializeBudgetItems(items)
			if err != nil {
				return nil, err
			}
			budgets = append(budgets, b)
		}
	}

	return budgets, nil
}

func (db *SqliteDB) UpdateBudget(id int, b Budget) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	items, err := serializeBudgetItems(b.Items)
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("UPDATE budgets SET year_month = (?), items = (?) WHERE id = (?);")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(b.YearMonth, items, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *SqliteDB) DeleteBudget(id int) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("DELETE FROM budgets WHERE id = (?);")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}

	return tx.Commit()
}
