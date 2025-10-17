package main

import (
	"fmt"
	"os"
	"tracker/market"
	"tracker/storage"
	"tracker/tui"
)

func main() {
	db, cleanup := storage.OpenLocalDatabase(false)
	defer cleanup()

	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "help":
			fmt.Println("Usage: tracker [help|update]")
			fmt.Println("  help: show this help")
			fmt.Println("  update: update market data")
			fmt.Println("  (no args): start the portfolio tracker app")
		case "update":
			market.UpdateMarketData(db)
			fmt.Println("Market data updated successfully")
		default:
			fmt.Println("Unknown command")
			fmt.Println("Usage: tracker [help|update]")
		}

		return
	} else if len(os.Args) > 2 {
		fmt.Println("Too many arguments")
		fmt.Println("Usage: tracker [help|update]")
		return
	}

	tui.StartApp(db)
}
