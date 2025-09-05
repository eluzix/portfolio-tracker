package types

import (
	"fmt"
	"time"
)

const formatYYYYMMDD = "2006-01-02"

type TransactionType string

const (
	TransactionTypeBuy TransactionType = "buy"
)

type Transaction struct {
	Id        string `json:"id"`
	AccountId string `json:"account_id"`
	Symbol    string `json:"symbol"`
	Date      string `json:"date"`
	Type      string `json:"transaction_type"`
	Quantity  uint32 `json:"quantity"`
	Pps       uint32 `json:"pps"`
}

func (t Transaction) AsDate() time.Time {
	ret, err := time.Parse(formatYYYYMMDD, t.Date)
	if err != nil {
		panic(fmt.Sprintf("Error formating date %s for transaction %s\n", t.Date, t.Id))
	}
	return ret
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
