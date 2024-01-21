package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Transaction struct {
	Id            int
	Date          string
	Description   string
	Amount        int
	AccountId     int
	UserId        int
	PlaidCategory string
	CategoryId    int
	PlaidId       string
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
	Password                string
}

type Account struct {
	Id             int
	UserId         int
	Name           string
	Balance        int
	PlaidAccountId string
}

type PlaidItem struct {
	Id                 int
	UserId             int
	AccessToken        string
	TransactionsCursor string
}

type Database struct {
	driver *sql.DB
}

func CreateDatabase() *Database {
	return &Database{}
}

func Create() (Database, error) {
	var db = Database{}
	var err error
	db.driver, err = sql.Open("sqlite3", "./db")
	if err != nil {
		return db, err
	}

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY,
		date TEXT,
		description TEXT,
		amount INT,
		account_id TEXT,
		user_id INT NOT NULL,
		plaid_category TEXT,
		category_id INT,
		plaid_id TEXT,
		FOREIGN KEY(user_id) REFERENCES users(id),
		FOREIGN KEY(account_id) REFERENCES accounts(id)
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

	CREATE TABLE IF NOT EXISTS plaid_items (
		id INTEGER PRIMARY KEY,
		user_id INT NOT NULL,
		access_token TEXT,
		transactions_cursor TEXT,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		name TEXT,
		password TEXT
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
	return db, err
}

func (db *Database) Close() error {
	return db.driver.Close()
}

func (db *Database) CreateUser(u User) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO users(name, password) VALUES(?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Name, u.Password)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *Database) GetUserByName(name string) (User, error) {
	var u User
	err := db.driver.QueryRow("SELECT * FROM users WHERE name = (?);", name).Scan(&u.Id, &u.Name, &u.Password)
	if err != nil {
		return u, err
	}

	return u, nil
}

func (db *Database) GetUserById(id int) (User, error) {
	var u User
	err := db.driver.QueryRow("SELECT * FROM users WHERE id = (?);", id).Scan(&u.Id, &u.Name, &u.Password)
	if err != nil {
		return u, err
	}

	return u, nil
}

func (db *Database) UpdateUser(id int, u User) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("UPDATE users SET name = (?), password = (?) WHERE id = (?);")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Name, u.Password, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *Database) DeleteUser(id int) error {
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

func (db *Database) CreateTransaction(t Transaction) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO transactions(date, description, amount, account_id, user_id, plaid_category, category_id, plaid_id) VALUES(?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(t.Date, t.Description, t.Amount, t.AccountId, t.UserId, t.PlaidCategory, t.CategoryId, t.PlaidId)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *Database) GetTransactions(user int) ([]Transaction, error) {
	rows, err := db.driver.Query("SELECT id, date, description, amount, account_id, user_id, plaid_category, category_id, plaid_id FROM transactions WHERE user_id = (?);", user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction

	for rows.Next() {
		var t Transaction
		err := rows.Scan(&t.Id, &t.Date, &t.Description, &t.Amount, &t.AccountId, &t.UserId, &t.PlaidCategory, &t.CategoryId, &t.PlaidId)
		if err != nil {
			return nil, err
		} else {
			transactions = append(transactions, t)
		}
	}

	return transactions, nil
}

func (db *Database) GetTransactionsDateRange(user int, start string, end string) ([]Transaction, error) {
	rows, err := db.driver.Query("SELECT id, date, description, amount, account_id, user_id, plaid_category, category_id, plaid_id FROM transactions WHERE user_id = (?) AND date >= (?) AND date <= (?);", user, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction

	for rows.Next() {
		var t Transaction
		err := rows.Scan(&t.Id, &t.Date, &t.Description, &t.Amount, &t.AccountId, &t.UserId, &t.PlaidCategory, &t.CategoryId, &t.PlaidId)
		if err != nil {
			return nil, err
		} else {
			transactions = append(transactions, t)
		}
	}

	return transactions, nil
}

func (db *Database) GetTransactionByPlaidId(user int, plaidId string) (Transaction, error) {
	var t Transaction
	err := db.driver.QueryRow("SELECT id, date, description, amount, account_id, user_id, plaid_category, category_id, plaid_id FROM transactions WHERE user_id = (?) AND plaid_id = (?);", user, plaidId).Scan(&t.Id, &t.Date, &t.Description, &t.Amount, &t.AccountId, &t.UserId, &t.PlaidCategory, &t.CategoryId, &t.PlaidId)
	if err != nil {
		return t, err
	}

	return t, nil
}

func (db *Database) UpdateTransaction(userId int, t Transaction) error {
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

func (db *Database) DeleteTransaction(id int) error {
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

func (db *Database) CreateCategory(c Category) error {
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

func (db *Database) GetCategories(user int) ([]Category, error) {
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

func (db *Database) UpdateCategory(userId int, c Category) error {
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

func (db *Database) DeleteCategory(id int) error {
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

func (db *Database) CreateAccount(a Account) error {
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

func (db *Database) GetAccounts(user int) ([]Account, error) {
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

func (db *Database) GetAccountByPlaidId(user int, plaidId string) (Account, error) {
	var a Account
	err := db.driver.QueryRow("SELECT * FROM accounts WHERE user_id = (?) AND plaid_account_id = (?);", user, plaidId).Scan(&a.Id, &a.UserId, &a.Name, &a.Balance, &a.PlaidAccountId)
	if err != nil {
		return a, err
	}

	return a, nil
}

func (db *Database) UpdateAccount(userId int, a Account) error {
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

func (db *Database) DeleteAccount(id int) error {
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

func (db *Database) CreatePlaidItem(item PlaidItem) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO plaid_items(user_id, access_token, transactions_cursor) VALUES(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(item.UserId, item.AccessToken, item.TransactionsCursor)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *Database) GetPlaidItems(user int) ([]PlaidItem, error) {
	rows, err := db.driver.Query("SELECT id, user_id, access_token, transactions_cursor FROM plaid_items WHERE user_id = (?);", user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []PlaidItem

	for rows.Next() {
		var item PlaidItem
		err := rows.Scan(&item.Id, &item.UserId, &item.AccessToken, &item.TransactionsCursor)
		if err != nil {
			return nil, err
		} else {
			items = append(items, item)
		}
	}

	return items, nil
}

func (db *Database) UpdatePlaidItem(userId int, item PlaidItem) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("UPDATE plaid_items SET access_token = (?), transactions_cursor = (?) WHERE user_id = (?) AND id = (?);")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(item.AccessToken, item.TransactionsCursor, userId, item.Id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *Database) DeletePlaidItem(id int) error {
	tx, err := db.driver.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("DELETE FROM plaid_items WHERE id = (?);")
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