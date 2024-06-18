package main

import (
	"fmt"
	"log"
	"math"
	"os"

	"strconv"
	"strings"

	"budget/database"
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

	PORT, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal("Error parsing PORT env variable:", err)
	}
	fmt.Println("Running at port", PORT)
	log.Fatal(api.Start(PORT, db))
}
