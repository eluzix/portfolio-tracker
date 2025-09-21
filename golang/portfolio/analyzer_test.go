package portfolio

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"testing"
	"time"
	"tracker/types"
	"tracker/utils"
)

func TestFirstLastTransaction(t *testing.T) {
	transactions := []types.Transaction{
		{Id: "id1", Date: utils.StringToDate("2025-01-01"), Type: types.TransactionTypeBuy, Symbol: "AAPL", Pps: 1, Quantity: 1},
		{Id: "id2", Date: utils.StringToDate("2025-02-01"), Type: types.TransactionTypeBuy, Symbol: "AAPL", Pps: 1, Quantity: 1},
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
			Date:      utils.StringToDate("2024-01-01"),
		},
		{
			AccountId: "1",
			Symbol:    "AAPL",
			Type:      types.TransactionTypeBuy,
			Quantity:  3,
			Pps:       100,
			Date:      utils.StringToDate("2024-02-01"),
		},
		{
			AccountId: "1",
			Symbol:    "AAPL",
			Type:      types.TransactionTypeBuy,
			Quantity:  1,
			Pps:       100,
			Date:      utils.StringToDate("2024-03-01"),
		},
		{
			AccountId: "1",
			Symbol:    "AAPL",
			Type:      types.TransactionTypeSell,
			Quantity:  4,
			Pps:       100,
			Date:      utils.StringToDate("2024-04-01"),
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
			Date:      utils.StringToDate("2024-01-01"),
		},
		{
			AccountId: "1",
			Symbol:    "AAPL",
			Type:      types.TransactionTypeBuy,
			Quantity:  3,
			Pps:       100,
			Date:      utils.StringToDate("2024-02-01"),
		},
		{
			AccountId: "1",
			Symbol:    "AAPL",
			Type:      types.TransactionTypeDividend,
			Quantity:  0,
			Pps:       10,
			Date:      utils.StringToDate("2024-02-15"),
		},
		{
			AccountId: "1",
			Symbol:    "AAPL",
			Type:      types.TransactionTypeSell,
			Quantity:  4,
			Pps:       100,
			Date:      utils.StringToDate("2024-05-01"),
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
			Date:     utils.StringToDate("2023-01-01"),
		},
		{
			Symbol:   "AAPL",
			Type:     types.TransactionTypeBuy,
			Quantity: int32(rand.Intn(95) + 5), // 5-100
			Pps:      1,
			Date:     utils.StringToDate("2023-02-01"),
		},
		{
			Symbol:   "AAPL",
			Type:     types.TransactionTypeBuy,
			Quantity: int32(rand.Intn(95) + 5), // 5-100
			Pps:      1,
			Date:     utils.StringToDate("2023-03-01"),
		},
		{
			Symbol:   "AAPL",
			Type:     types.TransactionTypeSell,
			Quantity: int32(rand.Intn(5) + 5), // 5-10
			Pps:      1,
			Date:     utils.StringToDate("2023-04-01"),
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
	dates := make([]time.Time, numTransactions)
	for i := 0; i < numTransactions; i++ {
		date := today.AddDate(-(i + 1), 0, 0) // Go back i+1 years
		dates[numTransactions-1-i] = date
	}

	firstTransactionDate := dates[0]
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
		transactionDate := t.Date
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

// =============================================================================
// FUZZY TESTING
// =============================================================================

// FuzzTestData holds generated test data for fuzzing
type FuzzTestData struct {
	transactions []types.Transaction
	priceTable   map[string]types.SymbolPrice
	description  string
}

// generateRandomTransactions creates random transactions for fuzzing
func generateRandomTransactions(seed int64, count int) []types.Transaction {
	r := rand.New(rand.NewSource(seed))
	transactions := make([]types.Transaction, count)

	symbols := []string{"AAPL", "GOOGL", "MSFT", "TSLA", "AMZN", "NVDA", "META"}
	transactionTypes := []types.TransactionType{
		types.TransactionTypeBuy,
		types.TransactionTypeSell,
		types.TransactionTypeDividend,
		types.TransactionTypeSplit,
	}

	baseDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < count; i++ {
		// Random date within last 4 years
		daysOffset := r.Intn(1461) // 4 years * 365.25 days
		date := baseDate.AddDate(0, 0, daysOffset)

		transactions[i] = types.Transaction{
			Id:        fmt.Sprintf("tx_%d", i),
			AccountId: fmt.Sprintf("acc_%d", r.Intn(5)+1),
			Symbol:    symbols[r.Intn(len(symbols))],
			Date:      date,
			Type:      transactionTypes[r.Intn(len(transactionTypes))],
			Quantity:  int32(r.Intn(1000) + 1), // 1-1000
			Pps:       int32(r.Intn(500) + 1),  // 1-500
		}
	}

	return transactions
}

// generateEdgeCaseTransactions creates edge case transactions for fuzzing
func generateEdgeCaseTransactions() []types.Transaction {
	return []types.Transaction{
		// Zero values
		{Id: "zero_qty", Symbol: "AAPL", Date: utils.StringToDate("2024-01-01"), Type: types.TransactionTypeBuy, Quantity: 0, Pps: 100},
		{Id: "zero_pps", Symbol: "AAPL", Date: utils.StringToDate("2024-01-01"), Type: types.TransactionTypeBuy, Quantity: 100, Pps: 0},

		// Maximum values
		{Id: "max_qty", Symbol: "AAPL", Date: utils.StringToDate("2024-01-01"), Type: types.TransactionTypeBuy, Quantity: math.MaxInt32, Pps: 1},
		{Id: "max_pps", Symbol: "AAPL", Date: utils.StringToDate("2024-01-01"), Type: types.TransactionTypeBuy, Quantity: 1, Pps: math.MaxInt32},

		// Edge dates
		{Id: "old_date", Symbol: "AAPL", Date: utils.StringToDate("1900-01-01"), Type: types.TransactionTypeBuy, Quantity: 100, Pps: 100},
		{Id: "future_date", Symbol: "AAPL", Date: utils.StringToDate("2100-12-31"), Type: types.TransactionTypeBuy, Quantity: 100, Pps: 100},
		{Id: "same_date_1", Symbol: "AAPL", Date: utils.StringToDate("2024-01-01"), Type: types.TransactionTypeBuy, Quantity: 100, Pps: 100},
		{Id: "same_date_2", Symbol: "AAPL", Date: utils.StringToDate("2024-01-01"), Type: types.TransactionTypeSell, Quantity: 50, Pps: 100},

		// Edge symbols
		{Id: "empty_symbol", Symbol: "", Date: utils.StringToDate("2024-01-01"), Type: types.TransactionTypeBuy, Quantity: 100, Pps: 100},
		{Id: "long_symbol", Symbol: strings.Repeat("A", 100), Date: utils.StringToDate("2024-01-01"), Type: types.TransactionTypeBuy, Quantity: 100, Pps: 100},
		{Id: "special_chars", Symbol: "AAPL!@#$%", Date: utils.StringToDate("2024-01-01"), Type: types.TransactionTypeBuy, Quantity: 100, Pps: 100},
		{Id: "unicode_symbol", Symbol: "AAPL🚀", Date: utils.StringToDate("2024-01-01"), Type: types.TransactionTypeBuy, Quantity: 100, Pps: 100},

		// Business logic edge cases
		{Id: "sell_before_buy", Symbol: "SELL", Date: utils.StringToDate("2024-01-01"), Type: types.TransactionTypeSell, Quantity: 100, Pps: 100},
		{Id: "dividend_no_shares", Symbol: "DIV", Date: utils.StringToDate("2024-01-01"), Type: types.TransactionTypeDividend, Quantity: 0, Pps: 10},
		{Id: "split_no_shares", Symbol: "SPLIT", Date: utils.StringToDate("2024-01-01"), Type: types.TransactionTypeSplit, Quantity: 0, Pps: 2},
	}
}

// generateRandomPriceTable creates random price table for fuzzing
func generateRandomPriceTable(seed int64, symbols []string) map[string]types.SymbolPrice {
	r := rand.New(rand.NewSource(seed))
	priceTable := make(map[string]types.SymbolPrice)

	for _, symbol := range symbols {
		if r.Float32() < 0.9 { // 90% chance to include symbol
			price := int32(r.Intn(1000) + 1) // 1-1000
			priceTable[symbol] = types.SymbolPrice{
				Symbol:   symbol,
				AdjPrice: price,
			}
		}
	}

	return priceTable
}

// extractUniqueSymbols gets unique symbols from transactions
func extractUniqueSymbols(transactions []types.Transaction) []string {
	symbolSet := make(map[string]bool)
	for _, t := range transactions {
		symbolSet[t.Symbol] = true
	}

	symbols := make([]string, 0, len(symbolSet))
	for symbol := range symbolSet {
		symbols = append(symbols, symbol)
	}
	return symbols
}

// validateNoPanic ensures the function doesn't panic
func validateNoPanic(t *testing.T, transactions []types.Transaction, priceTable map[string]types.SymbolPrice, testName string) (portfolio types.AnalyzedPortfolio, err error) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s: Function panicked with: %v", testName, r)
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	portfolio, err = AnalyzeTransactions(transactions, priceTable)
	return portfolio, err
}

// validateInvariants checks mathematical invariants
func validateInvariants(t *testing.T, transactions []types.Transaction, priceTable map[string]types.SymbolPrice, portfolio types.AnalyzedPortfolio, testName string) {
	var expectedInvested, expectedWithdrawn, expectedDividends int64
	symbolCounts := make(map[string]int32)

	for _, tx := range transactions {
		switch tx.Type {
		case types.TransactionTypeBuy:
			expectedInvested += int64(tx.Quantity * tx.Pps)
			symbolCounts[tx.Symbol] += tx.Quantity
		case types.TransactionTypeSell:
			expectedWithdrawn += int64(tx.Quantity * tx.Pps)
			symbolCounts[tx.Symbol] -= tx.Quantity
		case types.TransactionTypeDividend:
			count := symbolCounts[tx.Symbol]
			expectedDividends += int64(tx.Pps * count)
		case types.TransactionTypeSplit:
			symbolCounts[tx.Symbol] *= tx.Pps
		}
	}

	if portfolio.TotalInvested != expectedInvested {
		t.Errorf("%s: TotalInvested mismatch. Expected: %d, Got: %d", testName, expectedInvested, portfolio.TotalInvested)
	}

	if portfolio.TotalWithdrawn != expectedWithdrawn {
		t.Errorf("%s: TotalWithdrawn mismatch. Expected: %d, Got: %d", testName, expectedWithdrawn, portfolio.TotalWithdrawn)
	}

	if portfolio.TotalDividends != expectedDividends {
		t.Errorf("%s: TotalDividends mismatch. Expected: %d, Got: %d", testName, expectedDividends, portfolio.TotalDividends)
	}
}

// validateDeterminism ensures same input produces same output
func validateDeterminism(t *testing.T, transactions []types.Transaction, priceTable map[string]types.SymbolPrice, testName string) {
	portfolio1, err1 := AnalyzeTransactions(transactions, priceTable)
	portfolio2, err2 := AnalyzeTransactions(transactions, priceTable)

	if err1 != nil || err2 != nil {
		if err1 != err2 {
			t.Errorf("%s: Determinism failed - different errors: %v vs %v", testName, err1, err2)
		}
		return
	}

	if portfolio1 != portfolio2 {
		t.Errorf("%s: Determinism failed - different results", testName)
	}
}

// =============================================================================
// FUZZY TEST IMPLEMENTATIONS
// =============================================================================

func TestFuzzNoPanics(t *testing.T) {
	// Test with random data
	for i := 0; i < 100; i++ {
		seed := int64(i)
		testName := fmt.Sprintf("random_test_%d", i)

		// Generate random transactions
		transactionCount := rand.Intn(1000) + 1 // 1-1000 transactions
		transactions := generateRandomTransactions(seed, transactionCount)
		symbols := extractUniqueSymbols(transactions)
		priceTable := generateRandomPriceTable(seed, symbols)

		// Test for panics
		_, err := validateNoPanic(t, transactions, priceTable, testName)
		if err != nil && strings.Contains(err.Error(), "panic") {
			t.Errorf("%s failed with panic", testName)
		}
	}
}

func TestFuzzEdgeCases(t *testing.T) {
	edgeTransactions := generateEdgeCaseTransactions()
	symbols := extractUniqueSymbols(edgeTransactions)

	testCases := []struct {
		name       string
		priceTable map[string]types.SymbolPrice
	}{
		{
			name:       "edge_cases_with_prices",
			priceTable: generateRandomPriceTable(42, symbols),
		},
		{
			name:       "edge_cases_empty_prices",
			priceTable: make(map[string]types.SymbolPrice),
		},
		{
			name: "edge_cases_zero_prices",
			priceTable: map[string]types.SymbolPrice{
				"AAPL":  {Symbol: "AAPL", AdjPrice: 0},
				"SELL":  {Symbol: "SELL", AdjPrice: 0},
				"DIV":   {Symbol: "DIV", AdjPrice: 0},
				"SPLIT": {Symbol: "SPLIT", AdjPrice: 0},
			},
		},
		{
			name: "edge_cases_negative_prices",
			priceTable: map[string]types.SymbolPrice{
				"AAPL":  {Symbol: "AAPL", AdjPrice: -100},
				"SELL":  {Symbol: "SELL", AdjPrice: -50},
				"DIV":   {Symbol: "DIV", AdjPrice: -25},
				"SPLIT": {Symbol: "SPLIT", AdjPrice: -10},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := validateNoPanic(t, edgeTransactions, tc.priceTable, tc.name)
			if err == nil {
				validateDeterminism(t, edgeTransactions, tc.priceTable, tc.name)
			}
		})
	}
}

func TestFuzzInvariants(t *testing.T) {
	// Test mathematical invariants with various scenarios
	testCases := []struct {
		name  string
		setup func() ([]types.Transaction, map[string]types.SymbolPrice)
	}{
		{
			name: "simple_buy_sell",
			setup: func() ([]types.Transaction, map[string]types.SymbolPrice) {
				transactions := []types.Transaction{
					{Symbol: "AAPL", Date: utils.StringToDate("2024-01-01"), Type: types.TransactionTypeBuy, Quantity: 100, Pps: 150},
					{Symbol: "AAPL", Date: utils.StringToDate("2024-02-01"), Type: types.TransactionTypeSell, Quantity: 50, Pps: 160},
				}
				priceTable := map[string]types.SymbolPrice{
					"AAPL": {Symbol: "AAPL", AdjPrice: 170},
				}
				return transactions, priceTable
			},
		},
		{
			name: "multiple_symbols",
			setup: func() ([]types.Transaction, map[string]types.SymbolPrice) {
				transactions := generateRandomTransactions(123, 50)
				symbols := extractUniqueSymbols(transactions)
				priceTable := generateRandomPriceTable(123, symbols)
				return transactions, priceTable
			},
		},
		{
			name: "dividend_scenario",
			setup: func() ([]types.Transaction, map[string]types.SymbolPrice) {
				transactions := []types.Transaction{
					{Symbol: "AAPL", Date: utils.StringToDate("2024-01-01"), Type: types.TransactionTypeBuy, Quantity: 100, Pps: 150},
					{Symbol: "AAPL", Date: utils.StringToDate("2024-02-01"), Type: types.TransactionTypeDividend, Quantity: 0, Pps: 5},
					{Symbol: "AAPL", Date: utils.StringToDate("2024-03-01"), Type: types.TransactionTypeBuy, Quantity: 50, Pps: 160},
					{Symbol: "AAPL", Date: utils.StringToDate("2024-04-01"), Type: types.TransactionTypeDividend, Quantity: 0, Pps: 6},
				}
				priceTable := map[string]types.SymbolPrice{
					"AAPL": {Symbol: "AAPL", AdjPrice: 170},
				}
				return transactions, priceTable
			},
		},
		{
			name: "split_scenario",
			setup: func() ([]types.Transaction, map[string]types.SymbolPrice) {
				transactions := []types.Transaction{
					{Symbol: "AAPL", Date: utils.StringToDate("2024-01-01"), Type: types.TransactionTypeBuy, Quantity: 100, Pps: 200},
					{Symbol: "AAPL", Date: utils.StringToDate("2024-02-01"), Type: types.TransactionTypeSplit, Quantity: 0, Pps: 2},
					{Symbol: "AAPL", Date: utils.StringToDate("2024-03-01"), Type: types.TransactionTypeSell, Quantity: 50, Pps: 110},
				}
				priceTable := map[string]types.SymbolPrice{
					"AAPL": {Symbol: "AAPL", AdjPrice: 105},
				}
				return transactions, priceTable
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			transactions, priceTable := tc.setup()
			portfolio, err := validateNoPanic(t, transactions, priceTable, tc.name)
			if err == nil {
				validateInvariants(t, transactions, priceTable, portfolio, tc.name)
				validateDeterminism(t, transactions, priceTable, tc.name)
			}
		})
	}
}

func TestFuzzLargeDatasets(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset tests in short mode")
	}

	testSizes := []int{1000, 5000, 10000}

	for _, size := range testSizes {
		t.Run(fmt.Sprintf("large_dataset_%d", size), func(t *testing.T) {
			seed := int64(size)
			transactions := generateRandomTransactions(seed, size)
			symbols := extractUniqueSymbols(transactions)
			priceTable := generateRandomPriceTable(seed, symbols)

			// Measure performance
			start := time.Now()
			portfolio, err := validateNoPanic(t, transactions, priceTable, fmt.Sprintf("large_%d", size))
			duration := time.Since(start)

			if err == nil {
				t.Logf("Dataset size: %d, Duration: %v, Portfolio Value: %d", size, duration, portfolio.Value)

				// Basic sanity checks for large datasets
				if portfolio.TotalInvested < 0 {
					t.Errorf("TotalInvested should not be negative: %d", portfolio.TotalInvested)
				}
				if portfolio.TotalWithdrawn < 0 {
					t.Errorf("TotalWithdrawn should not be negative: %d", portfolio.TotalWithdrawn)
				}
			}
		})
	}
}

func TestFuzzMalformedDates(t *testing.T) {
	malformedDates := []string{
		"",
		"invalid-date",
		"2024-13-01", // Invalid month
		"2024-02-30", // Invalid day
		"20240101",   // Wrong format
		"2024/01/01", // Wrong separator
		"01-01-2024", // Wrong order
		"2024-1-1",   // No zero padding
	}

	for i, date := range malformedDates {
		t.Run(fmt.Sprintf("malformed_date_%d", i), func(t *testing.T) {
			// These should panic because of the panic in AsDate()
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Expected panic for malformed date %s but didn't get one", date)
				}
			}()

			transactions := []types.Transaction{
				{
					Id:       "test",
					Symbol:   "AAPL",
					Date:     utils.StringToDate(date),
					Type:     types.TransactionTypeBuy,
					Quantity: 100,
					Pps:      100,
				},
			}
			priceTable := map[string]types.SymbolPrice{
				"AAPL": {Symbol: "AAPL", AdjPrice: 100},
			}

			AnalyzeTransactions(transactions, priceTable)
		})
	}
}

func TestFuzzPropertyBased(t *testing.T) {
	// Property-based testing: test properties that should always hold
	for i := 0; i < 50; i++ {
		seed := int64(time.Now().UnixNano() + int64(i))
		r := rand.New(rand.NewSource(seed))

		// Generate realistic scenario
		transactionCount := r.Intn(100) + 10 // 10-110 transactions
		transactions := generateRandomTransactions(seed, transactionCount)
		symbols := extractUniqueSymbols(transactions)
		priceTable := generateRandomPriceTable(seed, symbols)

		testName := fmt.Sprintf("property_test_%d", i)
		portfolio, err := validateNoPanic(t, transactions, priceTable, testName)
		if err != nil {
			continue
		}

		// Property: Portfolio gain value should equal current value + withdrawn + dividends - invested
		expectedGain := (portfolio.Value + portfolio.TotalWithdrawn + portfolio.TotalDividends) - portfolio.TotalInvested
		if portfolio.GainValue != expectedGain {
			t.Errorf("%s: Portfolio gain calculation incorrect. Expected: %d, Got: %d",
				testName, expectedGain, portfolio.GainValue)
		}

		// Property: If no withdrawals or dividends, gain should be current value - invested
		if portfolio.TotalWithdrawn == 0 && portfolio.TotalDividends == 0 {
			expectedSimpleGain := portfolio.Value - portfolio.TotalInvested
			if portfolio.GainValue != expectedSimpleGain {
				t.Errorf("%s: Simple gain calculation incorrect. Expected: %d, Got: %d",
					testName, expectedSimpleGain, portfolio.GainValue)
			}
		}

		// Property: Determinism
		validateDeterminism(t, transactions, priceTable, testName)
	}
}
