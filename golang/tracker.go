package main

import (
	"tracker/market"
	"tracker/storage"
)

func main() {
	db, cleanup := storage.OpenLocalDatabase(false)
	defer cleanup()

	market.UpdateMarketData(db)
	// log := logging.Get()
	// log.Info("hello world\n")

}
