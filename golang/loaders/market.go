package loaders

import (
	"database/sql"
	"log/slog"
	"strings"
	"tracker/logging"
	"tracker/types"
)

func AllPrices(db *sql.DB) map[string]types.SymbolPrice {
	log := logging.Get()
	rows, err := db.Query("SELECT symbol, adj_close from prices")
	if err != nil {
		log.Error("failed to all transactions for user", slog.Any("error", err))
		return map[string]types.SymbolPrice{}
	}
	defer rows.Close()

	prices := make(map[string]types.SymbolPrice, 10)

	for rows.Next() {
		var p types.SymbolPrice
		_ = rows.Scan(&p.Symbol, &p.AdjPrice)
		prices[strings.ToLower(p.Symbol)] = p
	}

	return prices
}

func SymbolPrice(db *sql.DB, symbol string) types.SymbolPrice {
	log := logging.Get()
	var p types.SymbolPrice
	err := db.QueryRow("SELECT symbol, adj_close, created_at FROM prices WHERE symbol = ?", symbol).Scan(&p.Symbol, &p.AdjPrice, &p.CreatedAt)
	if err != nil {
		log.Error("failed to get price for symbol", slog.Any("error", err), slog.String("symbol", symbol))
		return types.SymbolPrice{}
	}
	return p
}

func CurrencyExchangeRate(db *sql.DB, symbol string) float64 {
	log := logging.Get()
	var v float64
	err := db.QueryRow("SELECT value FROM rates WHERE symbol = ?", symbol).Scan(&v)
	if err != nil {
		panic(err)
		log.Error("failed to get exchange rate for symbol", slog.Any("error", err), slog.String("symbol", symbol))
		return 1
	}
	return v
}
