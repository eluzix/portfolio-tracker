package portfolio

import (
	"sort"
	"strings"
	"tracker/types"
)

type TransactionRow struct {
	Transaction types.Transaction
	Quantity    int32
	Total       int64
}

func BuildTransactionRows(transactions []types.Transaction, showDividends bool) []TransactionRow {
	rows := make([]TransactionRow, 0, len(transactions))
	symbolsCount := make(map[string]int32)

	for _, tx := range transactions {
		symbol := strings.ToLower(tx.Symbol)
		quantity := tx.Quantity
		total := int64(tx.Quantity) * int64(tx.Pps)

		switch tx.Type {
		case types.TransactionTypeBuy:
			symbolsCount[symbol] += tx.Quantity
		case types.TransactionTypeSell:
			symbolsCount[symbol] -= tx.Quantity
		case types.TransactionTypeDividend:
			quantity = symbolsCount[symbol]
			total = int64(quantity) * int64(tx.Pps)
		case types.TransactionTypeSplit:
			count := symbolsCount[symbol]
			ratio := float32(tx.Pps) / 100
			symbolsCount[symbol] = int32(float32(count) * ratio)
		}

		rows = append(rows, TransactionRow{Transaction: tx, Quantity: quantity, Total: total})
	}

	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].Transaction.Date.After(rows[j].Transaction.Date)
	})

	if showDividends {
		return rows
	}

	filtered := make([]TransactionRow, 0, len(rows))
	for _, row := range rows {
		if row.Transaction.Type != types.TransactionTypeDividend && row.Transaction.Type != types.TransactionTypeSplit {
			filtered = append(filtered, row)
		}
	}

	return filtered
}
