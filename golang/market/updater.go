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
	UpdateMarketDataWithFetcher(db, NewMarketStackDataFetcher())
}

func UpdateMarketDataWithFetcher(db *sql.DB, fetcher DateFetcher) {
	start := time.Now()
	logger := logging.Get()
	logger.Info("Starting market data update")

	allSymbols, err := loaders.LoadAllSymbols(db)
	if err != nil {
		logger.Error("failed to load all symbols", slog.Any("error", err))
		return
	}
	logger.Info("Loaded symbols", slog.Int("count", len(allSymbols)))

	logger.Info("Fetching market data in parallel")
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
	logger.Info("Finished fetching market data",
		slog.Int("prices", len(prices)),
		slog.Int("dividends", len(dividends)),
		slog.Int("splits", len(splits)),
		slog.Int("rates", len(rates)))

	ctx := context.TODO()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("error creating transaction to update market info", slog.Any("error", err))
		return
	}
	defer tx.Rollback()

	now := time.Now()
	if prices != nil {
		logger.Info("Upserting prices", slog.Int("count", len(prices)))
		priceList := make([]types.SymbolPrice, 0, len(prices))
		for _, p := range prices {
			priceList = append(priceList, p)
		}
		if err := batchUpsertPrices(ctx, tx, priceList, now); err != nil {
			logger.Error("error batch upserting prices", slog.Any("error", err))
		}
		logger.Info("Finished upserting prices")
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
		logger.Info("Processing dividends/splits", slog.Int("count", len(allTransactions)))
		symbolSet := make(map[string]bool)
		for _, tr := range allTransactions {
			symbolSet[tr.Symbol] = true
		}

		symbols := make([]string, 0, len(symbolSet))
		for symbol := range symbolSet {
			symbols = append(symbols, symbol)
		}
		logger.Info("Deleting old dividends/splits", slog.Int("symbols", len(symbols)))
		if err := batchDeleteDividendsSplits(ctx, tx, symbols); err != nil {
			logger.Error("error batch deleting dividends/splits", slog.Any("error", err))
		}

		logger.Info("Inserting dividends/splits", slog.Int("count", len(allTransactions)))
		if err := batchInsertDividendsSplits(ctx, tx, allTransactions); err != nil {
			logger.Error("error batch inserting dividends/splits", slog.Any("error", err))
		}
		logger.Info("Finished processing dividends/splits")
	}

	if rates != nil {
		rateList := make([]struct {
			Symbol string
			Value  float64
		}, 0, len(rates))
		for symbol, value := range rates {
			if symbol == "USD" {
				continue
			}
			rateList = append(rateList, struct {
				Symbol string
				Value  float64
			}{symbol, value})
		}
		logger.Info("Upserting rates", slog.Int("count", len(rateList)))
		if err := batchUpsertRates(ctx, tx, rateList, now); err != nil {
			logger.Error("error batch upserting rates", slog.Any("error", err))
		}
		logger.Info("Finished upserting rates")
	}

	logger.Info("Committing transaction")
	err = tx.Commit()
	if err != nil {
		logger.Error("error committing transaction", slog.Any("error", err))
		return
	}

	logger.Info("Market data update completed", slog.Duration("duration", time.Since(start)))
}

const batchSize = 100

func batchUpsertPrices(ctx context.Context, tx *sql.Tx, prices []types.SymbolPrice, now time.Time) error {
	if len(prices) == 0 {
		return nil
	}

	const cols = 3
	for i := 0; i < len(prices); i += batchSize {
		end := min(i+batchSize, len(prices))
		batch := prices[i:end]

		placeholders := make([]byte, 0, len(batch)*(cols*2+3))
		for j := range batch {
			if j > 0 {
				placeholders = append(placeholders, ',')
			}
			placeholders = append(placeholders, "(?,?,?)"...)
		}

		query := "INSERT OR REPLACE INTO prices (symbol, adj_close, created_at) VALUES " + string(placeholders)
		args := make([]any, 0, len(batch)*cols)
		for _, p := range batch {
			args = append(args, p.Symbol, p.AdjPrice, now)
		}

		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}
	return nil
}

type rateEntry struct {
	Symbol string
	Value  float64
}

func batchUpsertRates(ctx context.Context, tx *sql.Tx, rates []struct{ Symbol string; Value float64 }, now time.Time) error {
	if len(rates) == 0 {
		return nil
	}

	const cols = 3
	for i := 0; i < len(rates); i += batchSize {
		end := min(i+batchSize, len(rates))
		batch := rates[i:end]

		placeholders := make([]byte, 0, len(batch)*(cols*2+3))
		for j := range batch {
			if j > 0 {
				placeholders = append(placeholders, ',')
			}
			placeholders = append(placeholders, "(?,?,?)"...)
		}

		query := "INSERT OR REPLACE INTO rates (symbol, value, created_at) VALUES " + string(placeholders)
		args := make([]any, 0, len(batch)*cols)
		for _, r := range batch {
			args = append(args, r.Symbol, r.Value, now)
		}

		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}
	return nil
}

func batchDeleteDividendsSplits(ctx context.Context, tx *sql.Tx, symbols []string) error {
	if len(symbols) == 0 {
		return nil
	}

	for i := 0; i < len(symbols); i += batchSize {
		end := min(i+batchSize, len(symbols))
		batch := symbols[i:end]

		placeholders := make([]byte, 0, len(batch)*2)
		for j := range batch {
			if j > 0 {
				placeholders = append(placeholders, ',')
			}
			placeholders = append(placeholders, '?')
		}

		query := "DELETE FROM dividends_splits WHERE symbol IN (" + string(placeholders) + ")"
		args := make([]any, len(batch))
		for j, s := range batch {
			args[j] = s
		}

		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}
	return nil
}

func batchInsertDividendsSplits(ctx context.Context, tx *sql.Tx, transactions []types.Transaction) error {
	if len(transactions) == 0 {
		return nil
	}

	const cols = 7
	for i := 0; i < len(transactions); i += batchSize {
		end := i + batchSize
		if end > len(transactions) {
			end = len(transactions)
		}
		batch := transactions[i:end]

		placeholders := make([]byte, 0, len(batch)*(cols*2+3))
		for j := range batch {
			if j > 0 {
				placeholders = append(placeholders, ',')
			}
			placeholders = append(placeholders, "(?,?,?,?,?,?,?)"...)
		}

		query := "INSERT INTO dividends_splits (id, account_id, symbol, date, transaction_type, quantity, pps) VALUES " + string(placeholders)
		args := make([]any, 0, len(batch)*cols)
		for _, tr := range batch {
			args = append(args,
				tr.Id,
				tr.AccountId,
				tr.Symbol,
				tr.Date.Format("2006-01-02"),
				tr.Type,
				tr.Quantity,
				tr.Pps,
			)
		}

		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}
	return nil
}
