package main

import (
	"fmt"
	"log/slog"
	"slices"
	"tracker/loaders"
	"tracker/logging"
	analyzer "tracker/portfolio"
	"tracker/storage"
	"tracker/types"
)

func main() {
	log := logging.Get()
	db, cleanup := storage.OpenLocalDatabase(false)
	defer cleanup()

	// market.UpdateMarketData(db)
	// log := logging.Get()
	// log.Info("hello world\n")

	// accounts, _ := loaders.UserAccounts(db)
	// log.Info(">>>>>accounts", slog.Any("accounsts", accounts))

	transactions, _ := loaders.AllTransactions(db)

	symbols := loaders.SymbolsFromTransactions(transactions)
	// log.Info(">>>>>symbols", slog.Any("symbols", symbols))

	firstTr := (*transactions)[0]
	dividends, _ := loaders.DividendsAndSplits(db, symbols, firstTr.Date)
	// log.Info(">>>>>dividends", slog.Any("dividends", dividends))

	allTransactions := append(*dividends, *transactions...)
	slices.SortFunc(allTransactions, func(a types.Transaction, b types.Transaction) int {
		ad := a.AsDate()
		bd := b.AsDate()

		if ad.After(bd) {
			return 1
		}

		if bd.After(ad) {
			return -1
		}

		return 0
	})

	// fmt.Printf(">>>>>>> tr1: %v\n", allTransactions[0])
	// fmt.Printf(">>>>>>> tr6: %d\n", len(allTransactions))

	prices := loaders.AllPrices(db)

	data, _ := analyzer.AnalyzeTransactions(allTransactions, prices)
	log.Info(">>>>> total", slog.Any("data", data))
	fmt.Printf("Total Value: %.2f\n", float64(data.Value)/100)
}
