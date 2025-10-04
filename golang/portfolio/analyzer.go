package portfolio

import (
	"fmt"
	"math"
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

	// symbolsValues := make(map[string]int64, len(pricesTable))
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
	var weigthedCashFlow int64 = 0

	for _, t := range transactions {
		symbol := strings.ToLower(t.Symbol)

		_, ok := pricesTable[symbol]
		if !ok {
			fmt.Printf("[AnalyzeTransactions] for %s missing price in table\n", t.Symbol)
			continue
		}

		trValue := int64(t.Quantity * t.Pps)
		daysSinceTransaction := int64(today.Sub(t.AsDate()).Hours() / 24)

		switch t.Type {
		case types.TransactionTypeBuy:
			totalInvested += trValue
			trCashFlow := trValue * (daysSinceInception - daysSinceTransaction) / daysSinceInception
			weigthedCashFlow += trCashFlow

			count, ok := symbolsCount[symbol]
			if !ok {
				count = 0
			}
			count += t.Quantity
			symbolsCount[symbol] = count

		case types.TransactionTypeSell:
			totalWithdrawn += trValue
			trCashFlow := trValue * (daysSinceInception - daysSinceTransaction) / daysSinceInception
			weigthedCashFlow -= trCashFlow

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
			dividendCashFlow := int64(trValue) * (daysSinceInception - daysSinceTransaction) / daysSinceInception
			weigthedCashFlow += dividendCashFlow

		case types.TransactionTypeSplit:
			count, ok := symbolsCount[symbol]
			if !ok {
				count = 0
			}

			pps := float32(t.Pps) / 100
			count = int32(float32(count) * pps)
			symbolsCount[symbol] = count
		}
	}

	for s, c := range symbolsCount {
		sp := pricesTable[s]
		portfolioValue += int64(sp.AdjPrice * c)
	}

	portfolio.Value = portfolioValue
	portfolio.TotalInvested = totalInvested
	portfolio.TotalWithdrawn = totalWithdrawn
	portfolio.TotalDividends = totalDividends

	portfolioGainValue := (portfolioValue + totalDividends + totalWithdrawn) - totalInvested
	portfolio.GainValue = portfolioGainValue
	if totalInvested == 0 {
		portfolio.Gain = 0
	} else {
		portfolio.Gain = float32(portfolioGainValue) / float32(totalInvested)
	}

	if weigthedCashFlow == 0 {
		portfolio.ModifiedDietzYield = 0
	} else {
		portfolio.ModifiedDietzYield = float32(float64(portfolioGainValue) / float64(totalInvested+weigthedCashFlow))
	}

	yearsSinceInception := float64(daysSinceInception) / 365
	portfolio.AnnualizedYield = float32(math.Pow(1+float64(portfolio.Gain), 1/yearsSinceInception)) - 1

	portfolio.SymbolsCount = symbolsCount
	portfolio.Transactions = transactions

	return portfolio, nil
}
