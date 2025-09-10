package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"
	"tracker/market"
	"tracker/storage"
	"tracker/types"
)

func LoadAllSymbols(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT DISTINCT(symbol) FROM transactions")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to all symbols: %s\n", err)
		return nil, err
	}
	defer rows.Close()

	symbols := make([]string, 0, 15)
	var symbol string
	for rows.Next() {
		err := rows.Scan(&symbol)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load next symbol: %s\n", err)
			return nil, err
		}

		symbols = append(symbols, symbol)
	}

	return symbols, nil
}

func UpdateMarketData(db *sql.DB) {
	allSymbols, err := LoadAllSymbols(db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to all symbols: %s\n", err)
		return
	}
	fmt.Printf("Done loading symbols, collected  %v\n", allSymbols)

	fetcher := market.NewMarketStackDataFetcher()
	var wg sync.WaitGroup

	var prices map[string]types.SymbolPrice
	// wg.Go(func() {
	// 	p, err := fetcher.FetchPrices(allSymbols)
	// 	if err != nil {
	// 		fmt.Printf("Error loading symbol prices: %s\n", err)
	// 		return
	// 	}
	//
	// 	prices = p
	// })

	var dividends map[string][]types.Transaction
	wg.Go(func() {
		d, err := fetcher.FetchDividends(allSymbols)
		if err != nil {
			fmt.Printf("Error loading symbols dividends: %s\n", err)
			return
		}

		dividends = d
	})

	var splits map[string][]types.Transaction
	wg.Go(func() {
		s, err := fetcher.FetchSplits(allSymbols)
		if err != nil {
			fmt.Printf("Error loading symbols splits: %s\n", err)
			return
		}

		splits = s
	})

	wg.Wait()

	ctx := context.TODO()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		fmt.Printf("error creating tx to update market info  >>>>>> %s\n", err)
		return
	}
	defer tx.Rollback()

	now := time.Now()
	if prices != nil {
		fmt.Printf("prices >>>>>> %v\n", prices)
		stat, err := tx.PrepareContext(ctx, "INSERT OR REPLACE INTO prices (symbol, adj_close, created_at) VALUES (?,?,?);")
		if err != nil {
			fmt.Printf("error creating statment to update prices >>>>>> %s\n", err)
		} else {
			defer stat.Close()

			for _, price := range prices {
				_, err := stat.Exec(price.Symbol, price.AdjPrice, now)
				if err != nil {
					fmt.Printf("error executing statment to update %s >>>>>> %s\n", price.Symbol, err)
				}
			}
		}
	}

	var allTransactions []types.Transaction = make([]types.Transaction, 0)
	for _, transactions := range dividends {
		for _, tr := range transactions {
			allTransactions = append(allTransactions, tr)
		}
	}
	for _, transactions := range splits {
		for _, tr := range transactions {
			allTransactions = append(allTransactions, tr)
		}
	}

	if len(allTransactions) > 0 {
		stat, err := tx.PrepareContext(ctx, "INSERT OR REPLACE INTO dividends_splits (id, account_id, symbol, date, transaction_type, quantity, pps) VALUES (?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			fmt.Printf("error creating statment to update dividends >>>>>> %s\n", err)
		} else {
			defer stat.Close()
			for _, tr := range allTransactions {
				_, err = stat.Exec(
					tr.Id,
					tr.AccountId,
					tr.Symbol,
					tr.Date,
					tr.Type,
					tr.Quantity,
					tr.Pps,
				)
				if err != nil {
					fmt.Printf("error executing statment to update %v >>>>>> %s\n", tr, err)
				}
			}
		}

	}

	err = tx.Commit()
	if err != nil {
		fmt.Printf("error commitging transaction >>>>>> %s\n", err)
	}
}

func main() {
	db, cleanup := storage.OpenLocalDatabase(false)
	defer cleanup()

	UpdateMarketData(db)

	// var count int
	// err := db.QueryRow("SELECT COUNT(*) FROM transactions").Scan(&count)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "failed to count transactions: %v\n", err)
	// 	os.Exit(1)
	// }
	//
	// fmt.Printf("Total transactions in database: %d\n", count)

}
