package loaders

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"tracker/logging"
	"tracker/types"
	"tracker/utils"
)

func AllTransactions(db *sql.DB) (*[]types.Transaction, error) {
	log := logging.Get()
	rows, err := db.Query("SELECT id,account_id,symbol,date,transaction_type,quantity,pps from transactions")
	if err != nil {
		log.Error("failed to all transactions for user", slog.Any("error", err))
		return nil, err
	}
	defer rows.Close()

	transactions := make([]types.Transaction, 0)
	for rows.Next() {
		var tr types.Transaction
		_ = rows.Scan(&tr.Id, &tr.AccountId, &tr.Symbol, &tr.Date, &tr.Type, &tr.Quantity, &tr.Pps)
		transactions = append(transactions, tr)
	}

	return &transactions, nil

}

func AccountTransactions(db *sql.DB, accountId string) (*[]types.Transaction, error) {
	log := logging.Get()
	rows, err := db.Query("SELECT id,account_id,symbol,date,transaction_type,quantity,pps from transactions WHERE account_id=?", accountId)
	if err != nil {
		log.Error("failed to all transactions for user account", slog.String("account", accountId), slog.Any("error", err))
		return nil, err
	}
	defer rows.Close()

	transactions := make([]types.Transaction, 0)
	for rows.Next() {
		var tr types.Transaction
		_ = rows.Scan(&tr.Id, &tr.AccountId, &tr.Symbol, &tr.Date, &tr.Type, &tr.Quantity, &tr.Pps)
		transactions = append(transactions, tr)
	}

	return &transactions, nil

}

func DividendsAndSplits(db *sql.DB, symbols []string, after time.Time) (*[]types.Transaction, error) {
	log := logging.Get()
	ph := make([]string, len(symbols))
	args := make([]any, len(symbols)+1)
	args[0] = after
	for i := range ph {
		ph[i] = "?"
		args[i+1] = symbols[i]
	}

	rows, err := db.Query(fmt.Sprintf("SELECT id,account_id,symbol,date,transaction_type,quantity,pps from dividends_splits where date > ? AND symbol in (%s)", strings.Join(ph, ",")), args...)
	if err != nil {
		log.Error("failed to all dividends and splits", slog.Any("error", err))
		return nil, err
	}
	defer rows.Close()

	transactions := make([]types.Transaction, 0)
	for rows.Next() {
		var tr types.Transaction
		_ = rows.Scan(&tr.Id, &tr.AccountId, &tr.Symbol, &tr.Date, &tr.Type, &tr.Quantity, &tr.Pps)
		transactions = append(transactions, tr)
	}

	return &transactions, nil

}

func AddTransaction(db *sql.DB, tr types.Transaction) error {
	if tr.Id == "" {
		tr.Id = utils.GenerateUUID()
	}
	_, err := db.Exec("insert into transactions (id,account_id,symbol,date,transaction_type,quantity,pps) values (?,?,?,?,?,?,?)", tr.Id, tr.AccountId, tr.Symbol, tr.Date, tr.Type, tr.Quantity, tr.Pps)

	return err
}

func DeleteTransaction(db *sql.DB, id string) error {
	if id == "" {
		return nil
	}

	_, err := db.Exec("delete from transactions where id=?", id)
	return err
}
