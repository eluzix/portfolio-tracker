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

	symbolsPrice := make(map[string]int32, 10)

	//todo add first and last transaction to portfolio
	firstTransaction := transactions[0]
	lastTransaction := transactions[totalTransactions-1]
	portfolio.FirstTransaction = firstTransaction
	portfolio.LastTransaction = lastTransaction

	// today := time.Now()
	// todo transform to days...
	// daysSinceInception := today.Sub(firstTransaction.AsDate()).Hours() / 24

	var totalInvested int64
	var totalWithdrawn int64
	var totalDividends int64
	var portfolioValue int64

	for _, t := range transactions {
		fmt.Printf(">>>> %+v\n", t)
		symbolValue := symbolsPrice[t.Symbol]
		if symbolValue == 0 {
			symbolValue = t.Pps
			symbolsPrice[t.Symbol] = symbolValue
		}

		// switch tp := t.Type; tp {
		trValue := int64(t.Quantity * t.Pps)
		switch t.Type {
		case types.TransactionTypeBuy:
			totalInvested += trValue
			portfolioValue += int64(t.Quantity) * trValue

		case types.TransactionTypeSell:
			totalWithdrawn += trValue
			portfolioValue -= int64(t.Quantity) * trValue

		case types.TransactionTypeDividend:
			totalDividends += trValue

		}
	}

	portfolio.Value = portfolioValue
	portfolio.TotalInvested = totalInvested
	portfolio.TotalWithdrawn = totalWithdrawn
	portfolio.TotalDividends = totalDividends

	return portfolio, nil
}
