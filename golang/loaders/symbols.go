package loaders

import (
	"database/sql"
	"fmt"
	"os"
	"tracker/types"
)

func LoadAllSymbols(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT DISTINCT(symbol) FROM transactions")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to all symbols: %s\n", err)
		return nil, err
	}
	defer rows.Close()

	symbols := make([]string, 0, 15)
	var symbol string
	for rows.Next() {
		err := rows.Scan(&symbol)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load next symbol: %s\n", err)
			return nil, err
		}

		symbols = append(symbols, symbol)
	}

	return symbols, nil
}

func SymbolsFromTransactions(transactions *[]types.Transaction) []string {

	symbols := make(map[string]struct{}, 0)
	for _, tr := range *transactions {
		symbols[tr.Symbol] = struct{}{}
	}

	ret := make([]string, 0, len(symbols))
	for k := range symbols {
		ret = append(ret, k)
	}

	return ret
}
