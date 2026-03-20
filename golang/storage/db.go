package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tursodatabase/go-libsql"
)

func removeReplicaWalFiles(dbPath string) {
	_ = os.Remove(dbPath + "-wal")
	_ = os.Remove(dbPath + "-shm")
}

func isWalInsertFrameError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "wal_insert_frame failed")
}

func configDataDir() string {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		panic(fmt.Sprintf("Unable to location config dir: %s", err))
	}

	dir := filepath.Join(cfgDir, "tracker", "data")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(fmt.Sprintf("Error creating config directory %s: %s", dir, err))
	}

	return dir
}

func tmpDataDir() string {
	dir, err := os.MkdirTemp("", "libsql-*")
	if err != nil {
		panic(fmt.Sprintf("Error creating temporary directory: %s\n", err))
	}

	err = os.Chmod(dir, 0744)
	if err != nil {
		panic(fmt.Sprintf("Error setting tmpdir permissions %s : %s\n", dir, err))
	}

	return dir
}

func OpenLocalDatabase(tmp bool) (*sql.DB, func()) {
	dbName := "tracker.db"
	var err error
	var dir string

	if tmp {
		// todo move from tmp directory to wellknown location so it will not be deleted
		dir, err = os.MkdirTemp("", "libsql-*")
		if err != nil {
			panic(fmt.Sprintf("Error creating temporary directory: %s\n", err))
		}

		err = os.Chmod(dir, 0744)
		if err != nil {
			panic(fmt.Sprintf("Error setting tmpdir permissions %s : %s\n", dir, err))
		}
	} else {
		dir = configDataDir()
	}

	dbPath := filepath.Join(dir, dbName)

	db, err := sql.Open("libsql", "file:"+dbPath)

	if err != nil {
		panic(fmt.Sprintf("Error creating db: %s\n", err))
	}

	// defer connector.Close()

	cleanup := func() {
		db.Close()
		if tmp {
			os.RemoveAll(dir)
		}
	}

	rows, err := db.Query("PRAGMA journal_mode = WAL; PRAGMA synchronous = NORMAL;")
	// rows, err := db.Query("PRAGMA journal_mode = WAL")
	if err != nil {
		cleanup()
		panic(fmt.Sprintf("error setting WAL mode: %s\n", err))
	}
	defer rows.Close()

	return db, cleanup
}

func OpenDatabase(tmp bool) (*sql.DB, func()) {
	dbName := "tracker.db"
	primaryUrl := os.Getenv("TRACKER_DATABASE_URL")
	authToken := os.Getenv("TRACKER_AUTH_TOKEN")

	// fmt.Printf("url: %s\n", primaryUrl)
	// fmt.Printf("token: %s\n", authToken)
	if primaryUrl == "" || authToken == "" {
		panic("Missing env vars: TRACKER_DATABASE_URL and TRACKER_AUTH_TOKEN")
	}

	dir := configDataDir()
	if tmp {
		dir = tmpDataDir()
	}
	dbPath := filepath.Join(dir, dbName)

	connector, err := libsql.NewEmbeddedReplicaConnector(dbPath, primaryUrl,
		libsql.WithAuthToken(authToken),
		libsql.WithReadYourWrites(true),
	)
	if err != nil && isWalInsertFrameError(err) {
		removeReplicaWalFiles(dbPath)
		connector, err = libsql.NewEmbeddedReplicaConnector(dbPath, primaryUrl,
			libsql.WithAuthToken(authToken),
			libsql.WithReadYourWrites(true),
		)
	}
	if err != nil {
		// fmt.Println("Error creating connector:", err)
		panic(fmt.Sprintf("Error creating connector: %s\n", err))
	}

	// defer connector.Close()

	db := sql.OpenDB(connector)
	// defer db.Close()

	cleanup := func() {
		connector.Close()
		db.Close()
		if tmp {
			os.RemoveAll(dir)
		}
	}

	rows, err := db.Query("PRAGMA journal_mode = WAL; PRAGMA synchronous = NORMAL;")
	// rows, err := db.Query("PRAGMA journal_mode = WAL")
	if err != nil {
		cleanup()
		panic(fmt.Sprintf("error setting WAL mode: %s\n", err))
	}
	defer rows.Close()

	return db, cleanup
}
