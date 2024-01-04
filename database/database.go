package database

import (
	"database/sql"

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
	GetTransactionsDateRange(user int, start string, end string) ([]Transaction, error)
	UpdateTransaction(userId int, t Transaction) error
	DeleteTransaction(id int) error
	// Categories
	CreateCategory(c Category) error
	GetCategories(user int) ([]Category, error)
	UpdateCategory(userId int, c Category) error
	DeleteCategory(id int) error
	// Accounts
	CreateAccount(a Account) error
	GetAccounts(user int) ([]Account, error)
	GetAccountByPlaidId(user int, plaidId string) (Account, error)
	UpdateAccount(userId int, a Account) error
	DeleteAccount(id int) error
}

type Transaction struct {
	Id            int
	Date          string
	Description   string
	Amount        int
	Account       string
	UserId        int
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

type Account struct {
	Id             int
	UserId         int
	Name           string
	Balance        int
	PlaidAccountId string
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
	if err != nil {
		return err
	}

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

	CREATE TABLE IF NOT EXISTS accounts (
		id INTEGER PRIMARY KEY,
		user_id INT NOT NULL,
		name TEXT,
		balance INT,
		plaid_account_id TEXT
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

func (db *SqliteDB) GetTransactionsDateRange(user int, start string, end string) ([]Transaction, error) {
	rows, err := db.driver.Query("SELECT * FROM transactions WHERE user_id = (?) AND date >= (?) AND date <= (?);", user, start, end)
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

func (db *SqliteDB) UpdateTransaction(userId int, t Transaction) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("UPDATE transactions SET date = (?), description = (?), amount = (?), category_id = (?) WHERE user_id = (?) AND id = (?);")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(t.Date, t.Description, t.Amount, t.CategoryId, userId, t.Id)
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

	stmt, err := tx.Prepare("INSERT INTO categories(name, user_id, available, budgeted) VALUES(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(c.Name, c.UserId, c.Available, c.Budgeted)
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
		var c Category
		err := rows.Scan(&c.Id, &c.Name, &c.UserId, &c.Available, &c.Budgeted)
		if err != nil {
			return nil, err
		} else {
			categories = append(categories, c)
		}
	}

	return categories, nil
}

func (db *SqliteDB) UpdateCategory(userId int, c Category) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("UPDATE categories SET name = (?), available = (?), budgeted = (?) WHERE user_id = (?) AND id = (?);")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(c.Name, c.Available, c.Budgeted, userId, c.Id)
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

func (db *SqliteDB) CreateAccount(a Account) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO accounts(user_id, name, balance, plaid_account_id) VALUES(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(a.UserId, a.Name, a.Balance, a.PlaidAccountId)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *SqliteDB) GetAccounts(user int) ([]Account, error) {
	rows, err := db.driver.Query("SELECT * FROM accounts WHERE user_id = (?);", user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []Account

	for rows.Next() {
		var a Account
		err := rows.Scan(&a.Id, &a.UserId, &a.Name, &a.Balance, &a.PlaidAccountId)
		if err != nil {
			return nil, err
		} else {
			accounts = append(accounts, a)
		}
	}

	return accounts, nil
}

func (db *SqliteDB) GetAccountByPlaidId(user int, plaidId string) (Account, error) {
	var a Account
	err := db.driver.QueryRow("SELECT * FROM accounts WHERE user_id = (?) AND plaid_account_id = (?);", user, plaidId).Scan(&a.Id, &a.UserId, &a.Name, &a.Balance, &a.PlaidAccountId)
	if err != nil {
		return a, err
	}

	return a, nil
}

func (db *SqliteDB) UpdateAccount(userId int, a Account) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("UPDATE accounts SET name = (?), balance = (?), plaid_account_id = (?) WHERE user_id = (?) AND id = (?);")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(a.Name, a.Balance, a.PlaidAccountId, userId, a.Id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *SqliteDB) DeleteAccount(id int) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("DELETE FROM accounts WHERE id = (?);")
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