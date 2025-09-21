package portfolio

import (
	"database/sql"
	"slices"
	"tracker/loaders"
	"tracker/types"
)

func LoadAndAnalyze(db *sql.DB, account types.Account) (types.AnalyzedPortfolio, error) {
	// log := logging.Get()
	var transactions *[]types.Transaction

	if account.Id != "" {
		transactions, _ = loaders.AccountTransactions(db, account.Id)
	} else {
		transactions, _ = loaders.AllTransactions(db)
	}

	symbols := loaders.SymbolsFromTransactions(transactions)
	// log.Info("Account>>>>", slog.Any("account", account))

	firstTr := (*transactions)[0]
	dividends, _ := loaders.DividendsAndSplits(db, symbols, firstTr.Date)

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

	prices := loaders.AllPrices(db)

	data, err := AnalyzeTransactions(allTransactions, prices)
	if err != nil {
		return types.AnalyzedPortfolio{}, err
	}

	return data, nil
}
