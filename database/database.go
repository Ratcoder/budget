package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func Create() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./db")
	if err != nil {
		return db, err
	}

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS transactions (
		transaction_id INTEGER PRIMARY KEY,
		year INTEGER NOT NULL,
		month INTEGER NOT NULL,
		day INTEGER NOT NULL,
		description TEXT,
		amount INTEGER NOT NULL,
		category_id INTEGER NOT NULL,
		account_id INTEGER NOT NULL,
		FOREIGN KEY(category_id) REFERENCES categories(category_id),
		FOREIGN KEY(account_id) REFERENCES accounts(account_id)
	);

	CREATE TABLE IF NOT EXISTS category_folders (
		category_folder_id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		user_id INTEGER NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(user_id)
	);

	CREATE TABLE IF NOT EXISTS categories (
		category_id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		category_folder_id INTEGER NOT NULL,
		FOREIGN KEY(category_folder_id) REFERENCES category_folders(category_folder_id),
	);

	CREATE TABLE IF NOT EXISTS assign (
		year INTEGER NOT NULL,
		month INTEGER NOT NULL,
		category_id INTEGER NOT NULL,
		amount INTEGER NOT NULL,
		PRIMARY KEY(year, month, category_id),
		FOREIGN KEY(category_id) REFERENCES categories(category_id)
	);

	CREATE TABLE IF NOT EXISTS amount_budgets (
		category_id INTEGER PRIMARY KEY,
		amount INTEGER NOT NULL,
		FOREIGN KEY(category_id) REFERENCES categories(category_id)
	);

	CREATE TABLE IF NOT EXISTS percent_budgets (
		category_id INTEGER PRIMARY KEY,
		percent INTEGER NOT NULL,
		FOREIGN KEY(category_id) REFERENCES categories(category_id)
	);

	CREATE TABLE IF NOT EXISTS users (
		user_id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		password TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS accounts (
		account_id INTEGER PRIMARY KEY,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(user_id)
	);
	-- PRAGMA foreign_keys = ON;
	`

	_, err = db.Exec(sqlStmt)
	return db, err
}