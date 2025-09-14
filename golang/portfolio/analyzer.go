package analyzer

import (
	"fmt"
	"strings"
	"time"
	"tracker/types"
)

func AnalyzeTransactions(transactions []types.Transaction, pricesTable map[string]types.SymbolPrice) (types.AnalyzedPortfolio, error) {
	portfolio := types.NewAnalyzedPortfolio()
	totalTransactions := len(transactions)
	if totalTransactions == 0 {
		return portfolio, nil
	}

	symbolsValues := make(map[string]int64, len(pricesTable))
	symbolsCount := make(map[string]int32, len(pricesTable))

	//todo add first and last transaction to portfolio
	firstTransaction := transactions[0]
	lastTransaction := transactions[totalTransactions-1]
	portfolio.FirstTransaction = firstTransaction
	portfolio.LastTransaction = lastTransaction

	today := time.Now()
	// todo transform to days...
	daysSinceInception := int64(today.Sub(firstTransaction.AsDate()).Hours() / 24)

	var totalInvested int64
	var totalWithdrawn int64
	var totalDividends int64
	var portfolioValue int64
	var weigthedCashFlow int64

	for _, t := range transactions {
		symbol := strings.ToLower(t.Symbol)

		symbolPrice, ok := pricesTable[symbol]
		if !ok {
			fmt.Printf("[AnalyzeTransactions] for %s missing price in table\n", t.Symbol)
			continue
		}

		symbolValue, ok := symbolsValues[symbol]
		if !ok {
			symbolValue = 0
		}

		trValue := int64(t.Quantity * t.Pps)
		daysSinceTransaction := int64(today.Sub(t.AsDate()).Hours() / 24)

		// switch tp := t.Type; tp {
		switch t.Type {
		case types.TransactionTypeBuy:
			totalInvested += trValue
			trCashFlow := trValue * daysSinceTransaction / daysSinceInception
			weigthedCashFlow += trCashFlow

			symbolValue += int64(t.Quantity * symbolPrice.AdjPrice)
			symbolsValues[symbol] = symbolValue

			count, ok := symbolsCount[symbol]
			if !ok {
				count = 0
			}
			count += t.Quantity
			symbolsCount[symbol] = count

		case types.TransactionTypeSell:
			totalWithdrawn += trValue
			trCashFlow := trValue * daysSinceTransaction / daysSinceInception
			weigthedCashFlow -= trCashFlow
			symbolValue -= int64(t.Quantity * symbolPrice.AdjPrice)
			symbolsValues[symbol] = symbolValue

			count, ok := symbolsCount[symbol]
			if !ok {
				count = 0
			}
			count -= t.Quantity
			symbolsCount[symbol] = count

		case types.TransactionTypeDividend:
			count, ok := symbolsCount[symbol]
			if !ok {
				count = 0
			}
			trValue := t.Pps * count
			totalDividends += int64(trValue)
			dividendCashFlow := trValue * int32(daysSinceInception) / int32(daysSinceTransaction)
			weigthedCashFlow -= int64(dividendCashFlow)

		case types.TransactionTypeSplit:
			count, ok := symbolsCount[symbol]
			if !ok {
				count = 0
			}
			count *= t.Pps
			symbolsCount[symbol] = count

			symbolValue -= int64(t.Quantity * t.Pps)
			symbolsValues[symbol] = symbolValue

		}
	}

	for _, value := range symbolsValues {
		portfolioValue += value
	}

	portfolio.Value = portfolioValue
	portfolio.TotalInvested = totalInvested
	portfolio.TotalWithdrawn = totalWithdrawn
	portfolio.TotalDividends = totalDividends

	portfolioGainValue := (portfolioValue + totalDividends + totalWithdrawn) - totalInvested
	portfolio.GainValue = portfolioGainValue
	if totalInvested == 0 {
		portfolio.GainValue = 0
	} else {
		portfolio.Gain = int32(portfolioGainValue / totalInvested)
	}

	if totalInvested+weigthedCashFlow == 0 {
		portfolio.ModifiedDietzYield = 0
	} else {
		portfolio.ModifiedDietzYield = int32(portfolioGainValue / (totalInvested + weigthedCashFlow))
	}

	return portfolio, nil
}
