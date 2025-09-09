package market

import (
	"errors"
	"tracker/types"
)

// MarketError represents an error that occurs during market data operations
type MarketError struct {
	Details string
}

func (e MarketError) Error() string {
	return "Error: " + e.Details
}

// DateFetcher interface defines methods for fetching market data
type DateFetcher interface {
	// FetchPrices fetches current prices for the given symbols
	// Returns a map of symbol to SymbolPrice, or nil if failed
	FetchPrices(symbols []string) (map[string]types.SymbolPrice, error)

	// FetchDividends fetches dividend transactions for the given symbols
	// Returns a map of symbol to slice of dividend transactions, or nil if failed
	FetchDividends(symbols []string) (map[string][]types.Transaction, error)

	// FetchSplits fetches stock split transactions for the given symbols
	// Returns a map of symbol to slice of split transactions, or nil if failed
	FetchSplits(symbols []string) (map[string][]types.Transaction, error)

	// FetchExchangeRates fetches current exchange rates
	// Returns a map of currency code to exchange rate (relative to USD)
	FetchExchangeRates() (map[string]float64, error)
}

// ErrMarketDataUnavailable is returned when market data cannot be fetched
var ErrMarketDataUnavailable = errors.New("market data unavailable")
