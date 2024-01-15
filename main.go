package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"

	"strconv"
	"strings"

	"budget/database"
	plaid "budget/plaid_connection"
	"budget/api"

	"github.com/joho/godotenv"
)

func DollarStringToCents(s string) (int, error) {
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, ",", "")
	amount, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return 0, err
	}
	return int(math.Round(amount * 100)), nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := database.Create()
	if err != nil {
		log.Fatal("Error opening database:", err)
	}
	defer db.Close()

	user, err := db.GetUserByName("test")
	if err != nil {
		log.Fatal("Error getting user:", err)
	}
	go plaid.SyncTransactions(user.Id, &db)
	go plaid.SyncAccounts(user.Id, &db)

	PORT, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal("Error parsing PORT env variable:", err)
	}
	fmt.Println("Running at port", PORT)
	log.Fatal(api.Start(PORT, &db))
}

func transactionsFromCSV(r io.Reader) ([]database.Transaction, error) {
	var transactions []database.Transaction
	csvReader := csv.NewReader(r)
	// Skip header
	csvReader.Read()
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return transactions, err
		}

		amount, err := DollarStringToCents(record[2])
		if err != nil {
			return transactions, err
		}

		transactions = append(transactions, database.Transaction{Date: record[0], Description: record[1], Amount: amount})
	}

	return transactions, nil
}
