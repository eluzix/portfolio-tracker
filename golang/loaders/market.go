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
