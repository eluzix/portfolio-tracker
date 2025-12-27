package main

import (
	"fmt"
	"os"
	"tracker/market"
	"tracker/storage"
	"tracker/tui"
	"tracker/web"
)

func main() {

	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "help":
			fmt.Println("Usage: tracker [help|update]")
			fmt.Println("  help: show this help")
			fmt.Println("  update: update market data")
			fmt.Println("  (no args): start the portfolio tracker app")
		case "update":
			db, cleanup := storage.OpenLocalDatabase(false)
			defer cleanup()
			market.UpdateMarketData(db)
			fmt.Println("Market data updated successfully")
		case "server":
			web.StartServer()
		default:
			fmt.Println("Unknown command")
			fmt.Println("Usage: tracker [help|update|server]")
		}

		return
	} else if len(os.Args) > 2 {
		fmt.Println("Too many arguments")
		fmt.Println("Usage: tracker [help|update]")
		return
	}

	db, cleanup := storage.OpenLocalDatabase(false)
	defer cleanup()
	tui.StartApp(db)
}
