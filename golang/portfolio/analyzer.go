package analyzer

import (
	"fmt"
	"tracker/types"
)

func AnalyzeTransactions(transactions []types.Transaction) (types.AnalyzedPortfolio, error) {
	portfolio := types.NewAnalyzedPortfolio()
	totalTransactions := len(transactions)
	if totalTransactions == 0 {
		return portfolio, nil
	}

	//todo add first and last transaction to portfolio
	firstTransaction := transactions[0]
	lastTransaction := transactions[totalTransactions-1]
	portfolio.FirstTransaction = firstTransaction
	portfolio.LastTransaction = lastTransaction

	// today := time.Now()
	// todo transform to days...
	// daysSinceInception := today.Sub(firstTransaction.AsDate()).Hours() / 24

	for _, t := range transactions {
		fmt.Printf(">>>> %+v\n", t)
	}

	return portfolio, nil
}
