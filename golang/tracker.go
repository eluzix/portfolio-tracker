package main

import (
	"log/slog"
	"tracker/loaders"
	"tracker/logging"
	"tracker/storage"
)

func main() {
	log := logging.Get()
	db, cleanup := storage.OpenLocalDatabase(false)
	defer cleanup()

	// market.UpdateMarketData(db)
	// log := logging.Get()
	// log.Info("hello world\n")

	accounts, _ := loaders.UserAccounts(db)
	log.Info(">>>>>accounts", slog.Any("accounsts", accounts))

}
