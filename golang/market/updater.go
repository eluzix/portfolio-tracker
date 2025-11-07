package market

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"time"
	"tracker/loaders"
	"tracker/logging"
	"tracker/types"
)

func UpdateMarketData(db *sql.DB) {
	logger := logging.Get()
	allSymbols, err := loaders.LoadAllSymbols(db)
	if err != nil {
		logger.Error("failed to load all symbols", slog.Any("error", err))
		return
	}
	logger.Info("Done loading symbols, collected", slog.Any("error", err))

	fetcher := NewMarketStackDataFetcher()
	var wg sync.WaitGroup

	var prices map[string]types.SymbolPrice
	wg.Go(func() {
		p, err := fetcher.FetchPrices(allSymbols)
		if err != nil {
			logger.Error("Error loading symbol prices", slog.Any("error", err))
			return
		}

		prices = p
	})

	var dividends map[string][]types.Transaction
	wg.Go(func() {
		d, err := fetcher.FetchDividends(allSymbols)
		if err != nil {
			logger.Error("Error loading symbols dividends", slog.Any("error", err))
			return
		}

		dividends = d
	})

	var splits map[string][]types.Transaction
	wg.Go(func() {
		s, err := fetcher.FetchSplits(allSymbols)
		if err != nil {
			logger.Error("Error loading symbols splits", slog.Any("error", err))
			return
		}

		splits = s
	})

	var rates map[string]float64
	wg.Go(func() {
		e, err := fetcher.FetchExchangeRates()
		if err != nil {
			logger.Error("Error loading rates", slog.Any("error", err))
			return
		}

		rates = e
	})

	wg.Wait()

	ctx := context.TODO()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("error creating transaction to update market info", slog.Any("error", err))
		return
	}
	defer tx.Rollback()

	now := time.Now()
	if prices != nil {
		stat, err := tx.PrepareContext(ctx, "INSERT OR REPLACE INTO prices (symbol, adj_close, created_at) VALUES (?,?,?);")
		if err != nil {
			logger.Error("error creating statement to update prices", slog.Any("error", err))
		} else {
			defer stat.Close()

			for _, price := range prices {
				_, err := stat.Exec(price.Symbol, price.AdjPrice, now)
				if err != nil {
					logger.Error("error executing statement to update price", slog.String("symbol", price.Symbol), slog.Any("error", err))
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
			logger.Error("error creating statement to update dividends", slog.Any("error", err))
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
					logger.Error("error executing statement to update transaction", slog.Any("transaction", tr), slog.Any("error", err))
				}
			}
		}
	}

	if rates != nil {
		stat, err := tx.PrepareContext(ctx, "INSERT OR REPLACE INTO rates (symbol, value, created_at) VALUES (?, ?, ?)")
		// TODO add rates history to DB

		if err != nil {
			logger.Error("error creating statement to update rates", slog.Any("error", err))
		} else {
			defer stat.Close()
		}
		for symbol, value := range rates {
			if symbol == "USD" {
				continue
			}

			_, err = stat.Exec(
				symbol,
				value,
				now,
			)

			if err != nil {
				logger.Error("error executing statement to update transaction", slog.Any("symbol", symbol), slog.Any("error", err))
			}
		}

	}

	err = tx.Commit()
	if err != nil {
		logger.Error("error committing transaction", slog.Any("error", err))
	}
}
