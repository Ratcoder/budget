package plaid_connection

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	// "log"
	"net/http"
	"os"
	"strconv"

	"budget/database"
)

var PLAID_CLIENT_ID string
var PLAID_SECRET string
var PLAID_ENV string

func loadEnv() {
	if PLAID_CLIENT_ID != "" && PLAID_SECRET != "" {
		return
	}
	PLAID_CLIENT_ID = os.Getenv("PLAID_CLIENT_ID")
	PLAID_SECRET = os.Getenv("PLAID_SECRET")
	PLAID_ENV = os.Getenv("PLAID_ENV")
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
		Products:     []string{"transactions"},
		CountryCodes: []string{"US"},
		Language:     "en",
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	reqReader := bytes.NewReader([]byte(reqBytes))

	// Send request
	resp, err := http.Post("https://" + PLAID_ENV + ".plaid.com/link/token/create", "application/json", reqReader)
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
	resp, err := http.Post("https://" + PLAID_ENV + ".plaid.com/item/public_token/exchange", "application/json", reqReader)
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

func SyncItems(userId int, db *database.Database) error {
	loadEnv()

	// Load plaid items from database
	items, err := (*db).GetPlaidItems(userId)
	if err != nil {
		return err
	}

	// Sync each plaid item
	for _, item := range items {
		err := syncAccounts(item, db)
		if err != nil {
			return err
		}
		err = syncTransactions(item, db)
		if err != nil {
			return err
		}
	}

	return nil
}

func syncTransactions(item database.PlaidItem, db *database.Database) error {
	// Prepare request
	req := struct {
		ClientID    string `json:"client_id"`
		Secret      string `json:"secret"`
		AccessToken string `json:"access_token"`
		Cursor      string `json:"cursor"`
	}{
		ClientID:    PLAID_CLIENT_ID,
		Secret:      PLAID_SECRET,
		AccessToken: item.AccessToken,
		Cursor:      item.TransactionsCursor,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return err
	}
	reqReader := bytes.NewReader([]byte(reqBytes))

	// Send request
	resp, err := http.Post("https://" + PLAID_ENV + ".plaid.com/transactions/sync", "application/json", reqReader)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Plaid API error")
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
		return err
	}
	for _, transaction := range responce.Added {
		t := database.Transaction{
			Date:          transaction.Date,
			Amount:        int(transaction.Amount * -100),
			Account:       transaction.AccountID,
			Description:   transaction.Name,
			UserId:        item.UserId,
			PlaidCategory: transaction.Category.Primary,
			PlaidId:       transaction.TransactionID,
		}
		err := (*db).CreateTransaction(t)
		if err != nil {
			return err
		}
	}

	for _, transaction := range responce.Modified {
		t, err := (*db).GetTransactionByPlaidId(item.UserId, transaction.TransactionID)
		if err != nil {
			continue
		}
		t.Date = transaction.Date
		t.Amount = int(transaction.Amount * -100)
		t.Account = transaction.AccountID
		t.Description = transaction.Name
		t.PlaidCategory = transaction.Category.Primary
		err = (*db).UpdateTransaction(item.UserId, t)
		if err != nil {
			return err
		}
	}

	for _, transaction := range responce.Removed {
		t, err := (*db).GetTransactionByPlaidId(item.UserId, transaction.TransactionID)
		if err != nil {
			continue
		}
		err = (*db).DeleteTransaction(t.Id)
		if err != nil {
			return err
		}
	}

	item.TransactionsCursor = responce.NextCursor
	err = (*db).UpdatePlaidItem(item.UserId, item)
	if err != nil {
		return err
	}

	if responce.HasMore {
		return syncTransactions(item, db)
	}

	return nil
}

func syncAccounts(item database.PlaidItem, db *database.Database) error {
	// Prepare request
	req := struct {
		ClientID    string `json:"client_id"`
		Secret      string `json:"secret"`
		AccessToken string `json:"access_token"`
	}{
		ClientID:    PLAID_CLIENT_ID,
		Secret:      PLAID_SECRET,
		AccessToken: item.AccessToken,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return err
	}
	reqReader := bytes.NewReader([]byte(reqBytes))

	// Send request
	resp, err := http.Post("https://" + PLAID_ENV + ".plaid.com/accounts/balance/get", "application/json", reqReader)
	if err != nil {
		return err
	}

	// Parse response
	type Account struct {
		AccountID string `json:"account_id"`
		Name      string `json:"name"`
		Balances  struct {
			Available float64 `json:"available"`
			Current   float64 `json:"current"`
		} `json:"balances"`
		Type string `json:"type"`
	}
	responce := struct {
		Accounts []Account `json:"accounts"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(&responce)
	if err != nil {
		return err
	}
	log.Println(responce)

	for _, account := range responce.Accounts {
		if account.Type != "depository" {
			continue
		}

		// Check if account already exists
		a, err := (*db).GetAccountByPlaidId(item.UserId, account.AccountID)
		if err != nil {
			// Create account
			a = database.Account{
				UserId: item.UserId,
				Name:   account.Name,
				Balance: int(account.Balances.Current * 100),
				PlaidAccountId:  account.AccountID,
			}
			err = (*db).CreateAccount(a)
			continue
		}

		// Update account
		a.Name = account.Name
		a.Balance = int(account.Balances.Current * 100)
		err = (*db).UpdateAccount(item.UserId, a)
	}
	return nil
}