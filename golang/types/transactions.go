package types

import (
	"time"
)

const formatYYYYMMDD = "2006-01-02T00:00:00Z"

type TransactionType string

const (
	TransactionTypeBuy      TransactionType = "Buy"
	TransactionTypeSell     TransactionType = "Sell"
	TransactionTypeDividend TransactionType = "Dividend"
	TransactionTypeSplit    TransactionType = "Split"
)

type Transaction struct {
	Id        string          `json:"id"`
	AccountId string          `json:"account_id"`
	Symbol    string          `json:"symbol"`
	Date      time.Time       `json:"date"`
	Type      TransactionType `json:"transaction_type"`
	Quantity  int32           `json:"quantity"`
	Pps       int32           `json:"pps"`
}

func (t Transaction) AsDate() time.Time {
	return t.Date
	// ret, err := time.Parse(formatYYYYMMDD, t.Date)
	// if err != nil {
	// 	panic(fmt.Sprintf("Error formating date %s for transaction %s\n", t.Date, t.Id))
	// }
	// return ret
}

type Account struct {
	Id            string   `json:"id"`
	Name          string   `json:"name"`
	Owner         string   `json:"owner"`
	Institution   string   `json:"institution"`
	InstitutionId string   `json:"institution_id"`
	Description   *string  `json:"description"`
	Tags          []string `json:"tags"`
	CreatedAt     *string  `json:"created_at"`
	UpdatedAt     *string  `json:"updated_at"`
}
