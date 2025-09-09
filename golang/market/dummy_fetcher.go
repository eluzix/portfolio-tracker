package market

import "tracker/types"

// DummyMarketFetcher is a test implementation of DateFetcher that returns empty results
type DummyMarketFetcher struct{}

// NewDummyMarketFetcher creates a new DummyMarketFetcher instance
func NewDummyMarketFetcher() *DummyMarketFetcher {
	return &DummyMarketFetcher{}
}

// FetchPrices returns an empty map of symbol prices
func (d *DummyMarketFetcher) FetchPrices(symbols []string) (map[string]types.SymbolPrice, error) {
	return make(map[string]types.SymbolPrice), nil
}

// FetchDividends returns an empty map of dividend transactions
func (d *DummyMarketFetcher) FetchDividends(symbols []string) (map[string][]types.Transaction, error) {
	return make(map[string][]types.Transaction), nil
}

// FetchSplits returns an empty map of split transactions
func (d *DummyMarketFetcher) FetchSplits(symbols []string) (map[string][]types.Transaction, error) {
	return make(map[string][]types.Transaction), nil
}

// FetchExchangeRates returns an empty map of exchange rates
func (d *DummyMarketFetcher) FetchExchangeRates() (map[string]float64, error) {
	return make(map[string]float64), nil
}
