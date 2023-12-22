package plaid_connection

import (
	"bytes"
	"encoding/json"
	"fmt"
	// "log"
	"net/http"
	"os"
	"strconv"

	"budget/database"
)

var PLAID_CLIENT_ID string
var PLAID_SECRET string

func loadEnv() {
	if PLAID_CLIENT_ID != "" && PLAID_SECRET != "" {
		return
	}
	PLAID_CLIENT_ID = os.Getenv("PLAID_CLIENT_ID")
	PLAID_SECRET = os.Getenv("PLAID_SECRET")
}

func CreateLinkToken(userId int) (linkToken string, err error) {
	loadEnv()

	// Prepare request
	type ReqUser struct {
		ClientUserId string `json:"client_user_id"`
	}
	req := struct {
		ClientID     string   `json:"client_id"`
		Secret       string   `json:"secret"`
		ClientName   string   `json:"client_name"`
		User         ReqUser  `json:"user"`
		Products     []string `json:"products"`
		CountryCodes []string `json:"country_codes"`
		Language     string   `json:"language"`
	}{
		ClientID:   PLAID_CLIENT_ID,
		Secret:     PLAID_SECRET,
		ClientName: "Plaid Test App",
		User: ReqUser{
			ClientUserId: strconv.Itoa(userId),
		},
		Products:     []string{"transactions", "auth"},
		CountryCodes: []string{"US"},
		Language:     "en",
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	reqReader := bytes.NewReader([]byte(reqBytes))

	// Send request
	resp, err := http.Post("https://sandbox.plaid.com/link/token/create", "application/json", reqReader)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Plaid API error")
	}

	// Parse response
	responce := struct {
		LinkToken string `json:"link_token"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&responce)
	if err != nil {
		return "", err
	}

	return responce.LinkToken, nil
}

func ExchangePublicToken(publicToken string) (accessToken string, err error) {
	loadEnv()

	// Prepare request
	req := struct {
		ClientID    string `json:"client_id"`
		Secret      string `json:"secret"`
		PublicToken string `json:"public_token"`
	}{
		ClientID:    PLAID_CLIENT_ID,
		Secret:      PLAID_SECRET,
		PublicToken: publicToken,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	reqReader := bytes.NewReader([]byte(reqBytes))

	// Send request
	resp, err := http.Post("https://sandbox.plaid.com/item/public_token/exchange", "application/json", reqReader)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Plaid API error")
	}

	// Parse response
	responce := struct {
		AccessToken string `json:"access_token"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&responce)
	if err != nil {
		return "", err
	}

	return responce.AccessToken, nil
}

func SyncTransactions(userId int, db *database.Database) (ret string, err error) {
	loadEnv()

	// Load user from database
	user, err := (*db).GetUserById(userId)
	if err != nil {
		return "", err
	}

	// Prepare request
	req := struct {
		ClientID    string `json:"client_id"`
		Secret      string `json:"secret"`
		AccessToken string `json:"access_token"`
		Cursor      string `json:"cursor"`
	}{
		ClientID:    PLAID_CLIENT_ID,
		Secret:      PLAID_SECRET,
		AccessToken: user.PlaidItem,
		Cursor:      user.PlaidTransactionsCursor,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	reqReader := bytes.NewReader([]byte(reqBytes))

	// Send request
	resp, err := http.Post("https://sandbox.plaid.com/transactions/sync", "application/json", reqReader)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Plaid API error")
	}

	// Parse response
	type Category struct {
		Primary string `json:"primary"`
	}
	type Transaction struct {
		AccountID     string   `json:"account_id"`
		TransactionID string   `json:"transaction_id"`
		Amount        float64  `json:"amount"`
		Date          string   `json:"date"`
		Name          string   `json:"name"`
		Category      Category `json:"personal_finance_category"`
	}
	type RemovedTransaction struct {
		TransactionID string `json:"transaction_id"`
	}
	responce := struct {
		NextCursor string               `json:"next_cursor"`
		HasMore    bool                 `json:"has_more"`
		Added      []Transaction        `json:"added"`
		Modified   []Transaction        `json:"modified"`
		Removed    []RemovedTransaction `json:"removed"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&responce)
	if err != nil {
		return "", err
	}
	for _, transaction := range responce.Added {
		t := database.Transaction{
			Date:        transaction.Date,
			Amount:      int(transaction.Amount * -100),
			Account:     transaction.AccountID,
			Description: transaction.Name,
			UserId:      userId,
			PlaidCategory:    transaction.Category.Primary,
		}
		err := (*db).CreateTransaction(t)
		if err != nil {
			return "", err
		}
	}

	user.PlaidTransactionsCursor = responce.NextCursor
	err = (*db).UpdateUser(userId, user)
	if err != nil {
		return "", err
	}

	if responce.HasMore {
		return SyncTransactions(userId, db)
	}

	return "", nil
}
