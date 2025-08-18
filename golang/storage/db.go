package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tursodatabase/go-libsql"
)

func OpenDatabase() (*sql.DB, func()) {
	dbName := "tracker.db"
	primaryUrl := os.Getenv("TRACKER_DATABASE_URL")
	authToken := os.Getenv("TRACKER_AUTH_TOKEN")
	fmt.Printf("url: %s\n", primaryUrl)
	fmt.Printf("token: %s\n", authToken)
	if primaryUrl == "" || authToken == "" {
		panic("Missing env vars: TRACKER_DATABASE_URL and TRACKER_AUTH_TOKEN")
	}

	// todo move from tmp directory to wellknown location so it will not be deleted
	dir, err := os.MkdirTemp("", "libsql-*")
	if err != nil {
		panic(fmt.Sprintf("Error creating temporary directory: %s\n", err))
	}
	// defer os.RemoveAll(dir)

	dbPath := filepath.Join(dir, dbName)

	connector, err := libsql.NewEmbeddedReplicaConnector(dbPath, primaryUrl,
		libsql.WithAuthToken(authToken),
	)
	if err != nil {
		// fmt.Println("Error creating connector:", err)
		panic(fmt.Sprintf("Error creating connector: %s\n", err))
	}

	// defer connector.Close()

	db := sql.OpenDB(connector)
	// defer db.Close()

	cleanup := func() {
		os.RemoveAll(dir)
		connector.Close()
		db.Close()
	}

	return db, cleanup
}
