package main

import (
	"tracker/storage"
	"tracker/tui"
)

func main() {
	db, cleanup := storage.OpenLocalDatabase(false)
	defer cleanup()

	tui.StartApp(db)
}
