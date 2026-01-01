package market

import (
	"database/sql"
	"fmt"
	"testing"
	"time"
	"tracker/types"

	_ "github.com/tursodatabase/go-libsql"
)

type MockFetcher struct {
	Prices    map[string]types.SymbolPrice
	Dividends map[string][]types.Transaction
	Splits    map[string][]types.Transaction
	Rates     map[string]float64

	PricesErr    error
	DividendsErr error
	SplitsErr    error
	RatesErr     error
}

func (m *MockFetcher) FetchPrices(symbols []string) (map[string]types.SymbolPrice, error) {
	if m.PricesErr != nil {
		return nil, m.PricesErr
	}
	return m.Prices, nil
}

func (m *MockFetcher) FetchDividends(symbols []string) (map[string][]types.Transaction, error) {
	if m.DividendsErr != nil {
		return nil, m.DividendsErr
	}
	return m.Dividends, nil
}

func (m *MockFetcher) FetchSplits(symbols []string) (map[string][]types.Transaction, error) {
	if m.SplitsErr != nil {
		return nil, m.SplitsErr
	}
	return m.Splits, nil
}

func (m *MockFetcher) FetchExchangeRates() (map[string]float64, error) {
	if m.RatesErr != nil {
		return nil, m.RatesErr
	}
	return m.Rates, nil
}

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	db, err := sql.Open("libsql", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}

	schemas := []string{
		`CREATE TABLE IF NOT EXISTS transactions (
			id TEXT PRIMARY KEY,
			account_id TEXT NOT NULL,
			symbol TEXT NOT NULL,
			date TEXT NOT NULL,
			transaction_type TEXT NOT NULL,
			quantity INTEGER NOT NULL,
			pps INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS dividends_splits (
			id TEXT PRIMARY KEY,
			account_id TEXT NOT NULL,
			symbol TEXT NOT NULL,
			date TEXT NOT NULL,
			transaction_type TEXT NOT NULL,
			quantity INTEGER NOT NULL,
			pps INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS accounts (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			owner TEXT NOT NULL,
			institution TEXT NOT NULL,
			institution_id TEXT NOT NULL,
			description TEXT,
			tags TEXT,
			created_at TEXT,
			updated_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS prices (
			symbol TEXT PRIMARY KEY,
			adj_close INTEGER NOT NULL,
			created_at DATETIME NULL
		)`,
		`CREATE TABLE IF NOT EXISTS rates (
			symbol TEXT PRIMARY KEY,
			value FLOAT NOT NULL,
			created_at DATETIME NULL
		)`,
	}

	for _, schema := range schemas {
		if _, err := db.Exec(schema); err != nil {
			db.Close()
			t.Fatalf("failed to create schema: %v", err)
		}
	}

	_, err = db.Exec(`INSERT INTO transactions (id, account_id, symbol, date, transaction_type, quantity, pps) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"tx1", "acc1", "AAPL", "2024-01-01", "Buy", 100, 15000)
	if err != nil {
		db.Close()
		t.Fatalf("failed to insert test transaction: %v", err)
	}

	cleanup := func() {
		db.Close()
	}
	return db, cleanup
}

func TestUpdateMarketData_Prices(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	fetcher := &MockFetcher{
		Prices: map[string]types.SymbolPrice{
			"AAPL": {Symbol: "AAPL", AdjPrice: 17500, CreatedAt: time.Now()},
			"GOOGL": {Symbol: "GOOGL", AdjPrice: 14000, CreatedAt: time.Now()},
		},
		Dividends: make(map[string][]types.Transaction),
		Splits:    make(map[string][]types.Transaction),
		Rates:     make(map[string]float64),
	}

	UpdateMarketDataWithFetcher(db, fetcher)

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM prices").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query prices count: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 prices, got %d", count)
	}

	var adjClose int
	err = db.QueryRow("SELECT adj_close FROM prices WHERE symbol = ?", "AAPL").Scan(&adjClose)
	if err != nil {
		t.Fatalf("failed to query AAPL price: %v", err)
	}
	if adjClose != 17500 {
		t.Errorf("expected AAPL price 17500, got %d", adjClose)
	}
}

func TestUpdateMarketData_DividendsAndSplits(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	testDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	fetcher := &MockFetcher{
		Prices: make(map[string]types.SymbolPrice),
		Dividends: map[string][]types.Transaction{
			"AAPL": {
				{Id: "div1", AccountId: "acc1", Symbol: "AAPL", Date: testDate, Type: types.TransactionTypeDividend, Quantity: 100, Pps: 50},
				{Id: "div2", AccountId: "acc1", Symbol: "AAPL", Date: testDate, Type: types.TransactionTypeDividend, Quantity: 100, Pps: 55},
			},
		},
		Splits: map[string][]types.Transaction{
			"AAPL": {
				{Id: "split1", AccountId: "acc1", Symbol: "AAPL", Date: testDate, Type: types.TransactionTypeSplit, Quantity: 4, Pps: 0},
			},
		},
		Rates: make(map[string]float64),
	}

	UpdateMarketDataWithFetcher(db, fetcher)

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM dividends_splits").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query dividends_splits count: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 dividends_splits, got %d", count)
	}

	var divCount, splitCount int
	err = db.QueryRow("SELECT COUNT(*) FROM dividends_splits WHERE transaction_type = ?", types.TransactionTypeDividend).Scan(&divCount)
	if err != nil {
		t.Fatalf("failed to query dividend count: %v", err)
	}
	err = db.QueryRow("SELECT COUNT(*) FROM dividends_splits WHERE transaction_type = ?", types.TransactionTypeSplit).Scan(&splitCount)
	if err != nil {
		t.Fatalf("failed to query split count: %v", err)
	}
	if divCount != 2 {
		t.Errorf("expected 2 dividends, got %d", divCount)
	}
	if splitCount != 1 {
		t.Errorf("expected 1 split, got %d", splitCount)
	}
}

func TestUpdateMarketData_Rates(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	fetcher := &MockFetcher{
		Prices:    make(map[string]types.SymbolPrice),
		Dividends: make(map[string][]types.Transaction),
		Splits:    make(map[string][]types.Transaction),
		Rates: map[string]float64{
			"EUR": 0.92,
			"GBP": 0.79,
			"USD": 1.0,
		},
	}

	UpdateMarketDataWithFetcher(db, fetcher)

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM rates").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query rates count: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 rates (USD skipped), got %d", count)
	}

	var eurRate float64
	err = db.QueryRow("SELECT value FROM rates WHERE symbol = ?", "EUR").Scan(&eurRate)
	if err != nil {
		t.Fatalf("failed to query EUR rate: %v", err)
	}
	if eurRate != 0.92 {
		t.Errorf("expected EUR rate 0.92, got %f", eurRate)
	}
}

func TestUpdateMarketData_PricesUpdate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := db.Exec("INSERT INTO prices (symbol, adj_close, created_at) VALUES (?, ?, ?)", "AAPL", 15000, time.Now())
	if err != nil {
		t.Fatalf("failed to insert initial price: %v", err)
	}

	fetcher := &MockFetcher{
		Prices: map[string]types.SymbolPrice{
			"AAPL": {Symbol: "AAPL", AdjPrice: 18000, CreatedAt: time.Now()},
		},
		Dividends: make(map[string][]types.Transaction),
		Splits:    make(map[string][]types.Transaction),
		Rates:     make(map[string]float64),
	}

	UpdateMarketDataWithFetcher(db, fetcher)

	var adjClose int
	err = db.QueryRow("SELECT adj_close FROM prices WHERE symbol = ?", "AAPL").Scan(&adjClose)
	if err != nil {
		t.Fatalf("failed to query AAPL price: %v", err)
	}
	if adjClose != 18000 {
		t.Errorf("expected updated AAPL price 18000, got %d", adjClose)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM prices WHERE symbol = ?", "AAPL").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query prices count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 AAPL price row (replaced), got %d", count)
	}
}

func TestUpdateMarketData_DividendsSplitsReplacement(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	oldDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := db.Exec("INSERT INTO dividends_splits (id, account_id, symbol, date, transaction_type, quantity, pps) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"old_div", "acc1", "AAPL", oldDate.Format("2006-01-02"), types.TransactionTypeDividend, 50, 25)
	if err != nil {
		t.Fatalf("failed to insert old dividend: %v", err)
	}

	newDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	fetcher := &MockFetcher{
		Prices: make(map[string]types.SymbolPrice),
		Dividends: map[string][]types.Transaction{
			"AAPL": {
				{Id: "new_div", AccountId: "acc1", Symbol: "AAPL", Date: newDate, Type: types.TransactionTypeDividend, Quantity: 100, Pps: 50},
			},
		},
		Splits: make(map[string][]types.Transaction),
		Rates:  make(map[string]float64),
	}

	UpdateMarketDataWithFetcher(db, fetcher)

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM dividends_splits WHERE symbol = ?", "AAPL").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query AAPL dividends count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 AAPL dividend (old replaced), got %d", count)
	}

	var id string
	err = db.QueryRow("SELECT id FROM dividends_splits WHERE symbol = ?", "AAPL").Scan(&id)
	if err != nil {
		t.Fatalf("failed to query AAPL dividend id: %v", err)
	}
	if id != "new_div" {
		t.Errorf("expected new_div id, got %s", id)
	}
}

func TestUpdateMarketData_LargeBatchDividends(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	testDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	dividends := make([]types.Transaction, 250)
	for i := 0; i < 250; i++ {
		dividends[i] = types.Transaction{
			Id:        fmt.Sprintf("div%d", i),
			AccountId: "acc1",
			Symbol:    "AAPL",
			Date:      testDate,
			Type:      types.TransactionTypeDividend,
			Quantity:  int32(i + 1),
			Pps:       50,
		}
	}

	fetcher := &MockFetcher{
		Prices: make(map[string]types.SymbolPrice),
		Dividends: map[string][]types.Transaction{
			"AAPL": dividends,
		},
		Splits: make(map[string][]types.Transaction),
		Rates:  make(map[string]float64),
	}

	UpdateMarketDataWithFetcher(db, fetcher)

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM dividends_splits").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query dividends_splits count: %v", err)
	}
	if count != 250 {
		t.Errorf("expected 250 dividends_splits, got %d", count)
	}
}

func TestUpdateMarketData_AllDataTypes(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	testDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	fetcher := &MockFetcher{
		Prices: map[string]types.SymbolPrice{
			"AAPL":  {Symbol: "AAPL", AdjPrice: 17500, CreatedAt: time.Now()},
			"GOOGL": {Symbol: "GOOGL", AdjPrice: 14000, CreatedAt: time.Now()},
		},
		Dividends: map[string][]types.Transaction{
			"AAPL": {
				{Id: "div1", AccountId: "acc1", Symbol: "AAPL", Date: testDate, Type: types.TransactionTypeDividend, Quantity: 100, Pps: 50},
			},
		},
		Splits: map[string][]types.Transaction{
			"GOOGL": {
				{Id: "split1", AccountId: "acc1", Symbol: "GOOGL", Date: testDate, Type: types.TransactionTypeSplit, Quantity: 20, Pps: 0},
			},
		},
		Rates: map[string]float64{
			"EUR": 0.92,
			"GBP": 0.79,
		},
	}

	UpdateMarketDataWithFetcher(db, fetcher)

	var priceCount, divSplitCount, rateCount int
	db.QueryRow("SELECT COUNT(*) FROM prices").Scan(&priceCount)
	db.QueryRow("SELECT COUNT(*) FROM dividends_splits").Scan(&divSplitCount)
	db.QueryRow("SELECT COUNT(*) FROM rates").Scan(&rateCount)

	if priceCount != 2 {
		t.Errorf("expected 2 prices, got %d", priceCount)
	}
	if divSplitCount != 2 {
		t.Errorf("expected 2 dividends_splits, got %d", divSplitCount)
	}
	if rateCount != 2 {
		t.Errorf("expected 2 rates, got %d", rateCount)
	}
}
