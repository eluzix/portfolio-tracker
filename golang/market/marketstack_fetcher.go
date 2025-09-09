package market

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"tracker/types"
)

// MarketStackDataFetcher implements DateFetcher using MarketStack and Exchange Rates APIs
type MarketStackDataFetcher struct {
	httpClient       *http.Client
	marketStackKey   string
	exchangeRatesKey string
}

// NewMarketStackDataFetcher creates a new MarketStackDataFetcher
func NewMarketStackDataFetcher() *MarketStackDataFetcher {
	return &MarketStackDataFetcher{
		httpClient:       &http.Client{},
		marketStackKey:   os.Getenv("MARKETSTACK_API_KEY"),
		exchangeRatesKey: os.Getenv("EXCHANGE_RATES_API_KEY"),
	}
}

// MarketStackResponse represents the response from MarketStack price API
type MarketStackResponse struct {
	Data []MarketStackPrice `json:"data"`
}

// MarketStackPrice represents a single price entry from MarketStack
type MarketStackPrice struct {
	Symbol   string  `json:"symbol"`
	AdjClose float64 `json:"adj_close"`
}

// MarketDividendsResponse represents the response from MarketStack dividends API
type MarketDividendsResponse struct {
	Data []MarketDividend `json:"data"`
}

// MarketDividend represents a single dividend entry from MarketStack
type MarketDividend struct {
	Date     string  `json:"date"`
	Dividend float64 `json:"dividend"`
	Symbol   string  `json:"symbol"`
}

// MarketSplitsResponse represents the response from MarketStack splits API
type MarketSplitsResponse struct {
	Data []MarketSplit `json:"data"`
}

// MarketSplit represents a single split entry from MarketStack
type MarketSplit struct {
	Date        string  `json:"date"`
	SplitFactor float64 `json:"split_factor"`
	Symbol      string  `json:"symbol"`
}

// ExchangeRatesResponse represents the response from Exchange Rates API
type ExchangeRatesResponse struct {
	Rates map[string]float64 `json:"rates"`
}

// FetchPrices fetches current prices from MarketStack API
func (m *MarketStackDataFetcher) FetchPrices(symbols []string) (map[string]types.SymbolPrice, error) {
	if m.marketStackKey == "" {
		return nil, fmt.Errorf("MARKETSTACK_API_KEY environment variable not set")
	}

	params := url.Values{}
	params.Add("symbols", strings.Join(symbols, ","))
	params.Add("access_key", m.marketStackKey)

	apiURL := "https://api.marketstack.com/v1/eod/latest?" + params.Encode()

	resp, err := m.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prices: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle the ,[] replacement from Rust version
	bodyStr := strings.ReplaceAll(string(body), ",[]", "")

	var marketResp MarketStackResponse
	if err := json.Unmarshal([]byte(bodyStr), &marketResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	prices := make(map[string]types.SymbolPrice, len(marketResp.Data))
	for _, price := range marketResp.Data {
		// Convert float64 to int32 (assuming price is in cents or similar)
		adjPriceInt := int32(price.AdjClose * 100)
		prices[price.Symbol] = types.SymbolPrice{
			Symbol:   price.Symbol,
			AdjPrice: adjPriceInt,
		}
	}

	return prices, nil
}

// FetchDividends fetches dividend transactions from MarketStack API
func (m *MarketStackDataFetcher) FetchDividends(symbols []string) (map[string][]types.Transaction, error) {
	if m.marketStackKey == "" {
		return nil, fmt.Errorf("MARKETSTACK_API_KEY environment variable not set")
	}

	params := url.Values{}
	params.Add("symbols", strings.Join(symbols, ","))
	params.Add("access_key", m.marketStackKey)
	params.Add("limit", "1000")

	apiURL := "https://api.marketstack.com/v1/dividends?" + params.Encode()

	resp, err := m.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch dividends: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var dividendsResp MarketDividendsResponse
	if err := json.Unmarshal(body, &dividendsResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	dividends := make(map[string][]types.Transaction)
	for _, div := range dividendsResp.Data {
		// Convert dividend amount to int32 (assuming in cents)
		ppsInt := int32(div.Dividend * 100)

		transaction := types.Transaction{
			Id:        "",
			AccountId: "",
			Symbol:    div.Symbol,
			Date:      div.Date,
			Type:      types.TransactionTypeDividend,
			Quantity:  0,
			Pps:       ppsInt,
		}

		dividends[div.Symbol] = append(dividends[div.Symbol], transaction)
	}

	return dividends, nil
}

// FetchSplits fetches stock split transactions from MarketStack API
func (m *MarketStackDataFetcher) FetchSplits(symbols []string) (map[string][]types.Transaction, error) {
	if m.marketStackKey == "" {
		return nil, fmt.Errorf("MARKETSTACK_API_KEY environment variable not set")
	}

	params := url.Values{}
	params.Add("symbols", strings.Join(symbols, ","))
	params.Add("access_key", m.marketStackKey)
	params.Add("limit", "1000")

	apiURL := "https://api.marketstack.com/v1/splits?" + params.Encode()

	resp, err := m.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch splits: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var splitsResp MarketSplitsResponse
	if err := json.Unmarshal(body, &splitsResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	splits := make(map[string][]types.Transaction)
	for _, split := range splitsResp.Data {
		// Convert split factor to int32 (multiply by 100 for precision)
		ppsInt := int32(split.SplitFactor * 100)

		transaction := types.Transaction{
			Id:        "",
			AccountId: "",
			Symbol:    split.Symbol,
			Date:      split.Date,
			Type:      types.TransactionTypeSplit,
			Quantity:  0,
			Pps:       ppsInt,
		}

		splits[split.Symbol] = append(splits[split.Symbol], transaction)
	}

	return splits, nil
}

// FetchExchangeRates fetches exchange rates from Exchange Rates API
func (m *MarketStackDataFetcher) FetchExchangeRates() (map[string]float64, error) {
	if m.exchangeRatesKey == "" {
		return nil, fmt.Errorf("EXCHANGE_RATES_API_KEY environment variable not set")
	}

	params := url.Values{}
	params.Add("base", "USD")
	params.Add("symbols", "ILS,EUR")

	apiURL := "https://api.apilayer.com/exchangerates_data/latest?" + params.Encode()

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("apikey", m.exchangeRatesKey)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exchange rates: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var exchangeResp ExchangeRatesResponse
	if err := json.Unmarshal(body, &exchangeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Add USD as 1.0 since it's the base currency
	if exchangeResp.Rates == nil {
		exchangeResp.Rates = make(map[string]float64)
	}
	exchangeResp.Rates["USD"] = 1.0

	return exchangeResp.Rates, nil
}
