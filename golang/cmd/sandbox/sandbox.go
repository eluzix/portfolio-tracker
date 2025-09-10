package main

import (
	"fmt"
	"log"
	"os"

	"tracker/market"
)

func main() {
	fmt.Println("MarketStack API Testing Sandbox")
	fmt.Println("==============================")

	// You can either:
	// 1. Set environment variables before running:
	//    export MARKETSTACK_API_KEY="your_key_here"
	//    export EXCHANGE_RATES_API_KEY="your_key_here"
	// 2. Or uncomment and set the keys directly below:
	
	// os.Setenv("MARKETSTACK_API_KEY", "your_marketstack_key_here")
	// os.Setenv("EXCHANGE_RATES_API_KEY", "your_exchange_rates_key_here")

	// Check if keys are set
	if os.Getenv("MARKETSTACK_API_KEY") == "" {
		fmt.Println("⚠️  MARKETSTACK_API_KEY not set")
		fmt.Println("   Either export it or uncomment the line in this file")
	}
	if os.Getenv("EXCHANGE_RATES_API_KEY") == "" {
		fmt.Println("⚠️  EXCHANGE_RATES_API_KEY not set")
		fmt.Println("   Either export it or uncomment the line in this file")
	}

	fetcher := market.NewMarketStackDataFetcher()
	symbols := []string{"AAPL"}

	fmt.Println("\n🍎 Testing with AAPL...")

	// Test prices
	fmt.Println("\n📈 Fetching current price...")
	if prices, err := fetcher.FetchPrices(symbols); err != nil {
		log.Printf("❌ Price fetch failed: %v", err)
	} else {
		fmt.Printf("✅ Fetched %d prices:\n", len(prices))
		for symbol, price := range prices {
			fmt.Printf("   %s: $%.2f\n", symbol, float64(price.AdjPrice)/100.0)
		}
	}

	// Test dividends
	fmt.Println("\n💰 Fetching dividends...")
	if dividends, err := fetcher.FetchDividends(symbols); err != nil {
		log.Printf("❌ Dividend fetch failed: %v", err)
	} else {
		fmt.Printf("✅ Fetched dividends for %d symbols:\n", len(dividends))
		for symbol, divs := range dividends {
			fmt.Printf("   %s: %d dividend payments\n", symbol, len(divs))
			for i, div := range divs {
				if i < 3 { // Show first 3
					fmt.Printf("     %s: $%.2f\n", div.Date, float64(div.Pps)/100.0)
				}
			}
			if len(divs) > 3 {
				fmt.Printf("     ... and %d more\n", len(divs)-3)
			}
		}
	}

	// Test splits
	fmt.Println("\n📊 Fetching stock splits...")
	if splits, err := fetcher.FetchSplits(symbols); err != nil {
		log.Printf("❌ Split fetch failed: %v", err)
	} else {
		fmt.Printf("✅ Fetched splits for %d symbols:\n", len(splits))
		for symbol, splitList := range splits {
			fmt.Printf("   %s: %d stock splits\n", symbol, len(splitList))
			for _, split := range splitList {
				fmt.Printf("     %s: %.2f:1 split\n", split.Date, float64(split.Pps)/100.0)
			}
		}
	}

	// Test exchange rates
	fmt.Println("\n💱 Fetching exchange rates...")
	if rates, err := fetcher.FetchExchangeRates(); err != nil {
		log.Printf("❌ Exchange rate fetch failed: %v", err)
	} else {
		fmt.Printf("✅ Fetched %d exchange rates:\n", len(rates))
		for currency, rate := range rates {
			fmt.Printf("   1 USD = %.4f %s\n", rate, currency)
		}
	}

	fmt.Println("\n🎉 Sandbox test completed!")
}
