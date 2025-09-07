package analyzer

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
	"tracker/types"
)

func TestFirstLastTransaction(t *testing.T) {
	transactions := []types.Transaction{
		{Id: "id1", Date: "2025-01-01", Type: types.TransactionTypeBuy, Symbol: "AAPL", Pps: 1, Quantity: 1},
		{Id: "id2", Date: "2025-02-01", Type: types.TransactionTypeBuy, Symbol: "AAPL", Pps: 1, Quantity: 1},
	}
	portfolio, err := AnalyzeTransactions(transactions, map[string]types.SymbolPrice{
		"AAPL": {AdjPrice: 12},
	})
	if err != nil {
		t.Fatalf("Error wasn't nil: %e\n", err)
	}

	if portfolio.FirstTransaction == (types.Transaction{}) {
		t.Fatalf("expected FirstTransaction to be NOT nil %e\n", err)
	}

	if portfolio.FirstTransaction.Id != "id1" {
		t.Fatalf("expected FirstTransaction.Id to be id1 but got %s\n", portfolio.FirstTransaction.Id)
	}

	if portfolio.LastTransaction.Id != "id2" {
		t.Fatalf("expected LastTransaction.Id to be id2 but got %s\n", portfolio.LastTransaction.Id)
	}
}

func TestEmptyTransactionsEmptyAnalyzer(t *testing.T) {
	priceTable := make(map[string]types.SymbolPrice)
	expected := types.NewAnalyzedPortfolio()
	
	portfolio, err := AnalyzeTransactions([]types.Transaction{}, priceTable)
	if err != nil {
		t.Fatalf("Error wasn't nil: %v\n", err)
	}

	if portfolio != expected {
		t.Fatalf("Expected empty portfolio but got %+v\n", portfolio)
	}
}

func TestTotals(t *testing.T) {
	transactions := []types.Transaction{
		{
			AccountId: "1",
			Symbol:    "AAPL",
			Type:      types.TransactionTypeBuy,
			Quantity:  2,
			Pps:       100,
			Date:      "2024-01-01",
		},
		{
			AccountId: "1",
			Symbol:    "AAPL",
			Type:      types.TransactionTypeBuy,
			Quantity:  3,
			Pps:       100,
			Date:      "2024-02-01",
		},
		{
			AccountId: "1",
			Symbol:    "AAPL",
			Type:      types.TransactionTypeBuy,
			Quantity:  1,
			Pps:       100,
			Date:      "2024-03-01",
		},
		{
			AccountId: "1",
			Symbol:    "AAPL",
			Type:      types.TransactionTypeSell,
			Quantity:  4,
			Pps:       100,
			Date:      "2024-04-01",
		},
	}

	priceTable := map[string]types.SymbolPrice{
		"AAPL": {Symbol: "AAPL", AdjPrice: 100},
	}

	portfolio, err := AnalyzeTransactions(transactions, priceTable)
	if err != nil {
		t.Fatalf("Error wasn't nil: %v\n", err)
	}

	if portfolio.TotalInvested != 600 {
		t.Fatalf("Expected TotalInvested to be 600 but got %d\n", portfolio.TotalInvested)
	}

	if portfolio.TotalWithdrawn != 400 {
		t.Fatalf("Expected TotalWithdrawn to be 400 but got %d\n", portfolio.TotalWithdrawn)
	}
}

func TestDividends(t *testing.T) {
	transactions := []types.Transaction{
		{
			AccountId: "1",
			Symbol:    "AAPL",
			Type:      types.TransactionTypeBuy,
			Quantity:  2,
			Pps:       100,
			Date:      "2024-01-01",
		},
		{
			AccountId: "1",
			Symbol:    "AAPL",
			Type:      types.TransactionTypeBuy,
			Quantity:  3,
			Pps:       100,
			Date:      "2024-02-01",
		},
		{
			AccountId: "1",
			Symbol:    "AAPL",
			Type:      types.TransactionTypeDividend,
			Quantity:  0,
			Pps:       10,
			Date:      "2024-02-15",
		},
		{
			AccountId: "1",
			Symbol:    "AAPL",
			Type:      types.TransactionTypeSell,
			Quantity:  4,
			Pps:       100,
			Date:      "2024-05-01",
		},
	}

	priceTable := map[string]types.SymbolPrice{
		"AAPL": {Symbol: "AAPL", AdjPrice: 100},
	}

	portfolio, err := AnalyzeTransactions(transactions, priceTable)
	if err != nil {
		t.Fatalf("Error wasn't nil: %v\n", err)
	}

	if portfolio.TotalInvested != 500 {
		t.Fatalf("Expected TotalInvested to be 500 but got %d\n", portfolio.TotalInvested)
	}

	if portfolio.TotalWithdrawn != 400 {
		t.Fatalf("Expected TotalWithdrawn to be 400 but got %d\n", portfolio.TotalWithdrawn)
	}

	if portfolio.TotalDividends != 50 {
		t.Fatalf("Expected TotalDividends to be 50 but got %d\n", portfolio.TotalDividends)
	}
}

func TestSymbolsValue(t *testing.T) {
	transactions := []types.Transaction{
		{
			Symbol:   "AAPL",
			Type:     types.TransactionTypeBuy,
			Quantity: int32(rand.Intn(95) + 5), // 5-100
			Pps:      1,
			Date:     "2023-01-01",
		},
		{
			Symbol:   "AAPL",
			Type:     types.TransactionTypeBuy,
			Quantity: int32(rand.Intn(95) + 5), // 5-100
			Pps:      1,
			Date:     "2023-02-01",
		},
		{
			Symbol:   "AAPL",
			Type:     types.TransactionTypeBuy,
			Quantity: int32(rand.Intn(95) + 5), // 5-100
			Pps:      1,
			Date:     "2023-03-01",
		},
		{
			Symbol:   "AAPL",
			Type:     types.TransactionTypeSell,
			Quantity: int32(rand.Intn(5) + 5), // 5-10
			Pps:      1,
			Date:     "2023-04-01",
		},
	}

	price := int32(rand.Intn(100) + 100) // 100-200
	var expectedValue int64 = 0
	for _, t := range transactions {
		if t.Type == types.TransactionTypeSell {
			expectedValue -= int64(t.Quantity * price)
		} else {
			expectedValue += int64(t.Quantity * price)
		}
	}

	priceTable := map[string]types.SymbolPrice{
		"AAPL": {Symbol: "AAPL", AdjPrice: price},
	}

	portfolio, err := AnalyzeTransactions(transactions, priceTable)
	if err != nil {
		t.Fatalf("Error wasn't nil: %v\n", err)
	}

	if fmt.Sprintf("%.5f", float64(portfolio.Value)) != fmt.Sprintf("%.5f", float64(expectedValue)) {
		t.Fatalf("Expected Value to be %d but got %d\n", expectedValue, portfolio.Value)
	}
}

func TestYields(t *testing.T) {
	today := time.Now()
	price := int32(rand.Intn(100) + 100) // 100-200

	numTransactions := 4
	quantities := []int32{
		int32(rand.Intn(95) + 5), // 5-100
		int32(rand.Intn(95) + 5), // 5-100
		int32(rand.Intn(95) + 5), // 5-100
		int32(rand.Intn(5) + 5),  // 5-10
	}
	transactionTypes := []types.TransactionType{
		types.TransactionTypeBuy,
		types.TransactionTypeBuy,
		types.TransactionTypeBuy,
		types.TransactionTypeSell,
	}

	// Create dates going back in time
	dates := make([]string, numTransactions)
	for i := 0; i < numTransactions; i++ {
		date := today.AddDate(-(i+1), 0, 0) // Go back i+1 years
		dates[numTransactions-1-i] = date.Format("2006-01-02")
	}

	firstTransactionDate, _ := time.Parse("2006-01-02", dates[0])
	daysSinceInception := int64(today.Sub(firstTransactionDate).Hours() / 24)

	transactions := make([]types.Transaction, numTransactions)
	for i := 0; i < numTransactions; i++ {
		transactions[i] = types.Transaction{
			Symbol:   "AAPL",
			Type:     transactionTypes[i],
			Quantity: quantities[i],
			Pps:      1,
			Date:     dates[i],
		}
	}

	var currentPortfolioValue int64
	var totalWithdrawn int64
	var totalInvested int64
	var weightedCashFlows int64

	for _, t := range transactions {
		transactionDate, _ := time.Parse("2006-01-02", t.Date)
		daysSinceTransaction := int64(today.Sub(transactionDate).Hours() / 24)
		
		switch t.Type {
		case types.TransactionTypeBuy:
			currentPortfolioValue += int64(t.Quantity * price)
			totalInvested += int64(t.Quantity * t.Pps)
			trCashFlow := int64(t.Quantity*t.Pps) * daysSinceTransaction / daysSinceInception
			weightedCashFlows += trCashFlow
		case types.TransactionTypeSell:
			currentPortfolioValue -= int64(t.Quantity * price)
			totalWithdrawn += int64(t.Quantity * t.Pps)
			trCashFlow := int64(t.Quantity*t.Pps) * daysSinceTransaction / daysSinceInception
			weightedCashFlows -= trCashFlow
		}
	}

	totalDividends := int64(0)
	portfolioGainValue := (currentPortfolioValue + totalWithdrawn + totalDividends) - totalInvested
	var expectedModifiedDietzYield int32
	if (totalInvested + weightedCashFlows) != 0 {
		expectedModifiedDietzYield = int32(portfolioGainValue / (totalInvested + weightedCashFlows))
	}

	priceTable := map[string]types.SymbolPrice{
		"AAPL": {Symbol: "AAPL", AdjPrice: price},
	}

	portfolio, err := AnalyzeTransactions(transactions, priceTable)
	if err != nil {
		t.Fatalf("Error wasn't nil: %v\n", err)
	}

	if portfolio.ModifiedDietzYield != expectedModifiedDietzYield {
		t.Fatalf("Expected ModifiedDietzYield to be %d but got %d\n", expectedModifiedDietzYield, portfolio.ModifiedDietzYield)
	}
}
