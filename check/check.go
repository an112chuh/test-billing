package check

import (
	"regexp"
	"work/config"
)

func IsUserValid(ID int, Email *string) (bool, error) {
	db := config.GetConnection()
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	params := []any{ID}
	if Email != nil {
		query = `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND email = $2)`
		params = []any{ID, &Email}
	}
	var exists bool
	err := db.QueryRow(query, params...).Scan(&exists)
	return exists, err
}

func IsBillValid(ID int) bool {
	db := config.GetConnection()
	query := `SELECT EXISTS(SELECT 1 FROM transactions WHERE id = $1)`
	params := []any{ID}
	var exists bool
	db.QueryRow(query, params...).Scan(&exists)
	return exists
}

func IsCurrencyValid(Currency string) bool {
	r, _ := regexp.Compile("[A-Z]{3}")
	return len(Currency) == 3 && r.MatchString(Currency)
}

func IsNewBillStatusValid(status int) bool {
	return status == 2 || status == 3
}

func IsApiKeyValid(apiKey string) (bool, error) {
	db := config.GetConnection()
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = 0 AND api_key = $1)`
	params := []any{apiKey}
	var exists bool
	err := db.QueryRow(query, params...).Scan(&exists)
	return exists, err
}

func CurrentStatus(ID int) (int, error) {
	db := config.GetConnection()
	query := `SELECT status FROM transactions WHERE id = $1`
	params := []any{ID}
	var status int
	err := db.QueryRow(query, params...).Scan(&status)
	return status, err
}
