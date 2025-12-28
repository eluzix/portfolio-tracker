package portfolio

import (
	"database/sql"
	"slices"
	"tracker/loaders"
	"tracker/types"
)

func LoadAndAnalyze(db *sql.DB, account types.Account) (types.AnalyzedPortfolio, error) {
	var transactions *[]types.Transaction

	if account.Id != "" {
		transactions, _ = loaders.AccountTransactions(db, account.Id)
	} else {
		transactions, _ = loaders.AllTransactions(db)
	}

	return analyzeTransactionSet(db, transactions)
}

func LoadAndAnalyzeAccounts(db *sql.DB, accountIds []string) (types.AnalyzedPortfolio, error) {
	if len(accountIds) == 0 {
		return types.AnalyzedPortfolio{}, nil
	}

	transactions, _ := loaders.AccountsTransactions(db, accountIds)
	return analyzeTransactionSet(db, transactions)
}

func analyzeTransactionSet(db *sql.DB, transactions *[]types.Transaction) (types.AnalyzedPortfolio, error) {
	if len(*transactions) == 0 {
		return types.AnalyzedPortfolio{}, nil
	}

	symbols := loaders.SymbolsFromTransactions(transactions)

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
