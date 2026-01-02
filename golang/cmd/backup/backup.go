package main

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	"tracker/storage"
)

func main() {
	if err := runBackup(); err != nil {
		fmt.Fprintf(os.Stderr, "Backup failed: %v\n", err)
		os.Exit(1)
	}
}

func runBackup() error {
	fmt.Println("=== Starting Database Backup ===")

	db, cleanup := storage.OpenDatabase()
	defer cleanup()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_150405")
	backupFileName := fmt.Sprintf("tracker_backup_%s.db", timestamp)
	backupPath := filepath.Join(homeDir, backupFileName)

	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}
	sourceDbPath := filepath.Join(cfgDir, "tracker", "data", "tracker.db")

	if err := copyDatabaseFile(sourceDbPath, backupPath); err != nil {
		return fmt.Errorf("failed to copy database: %w", err)
	}

	fmt.Printf("Database copied to: %s\n", backupPath)

	if err := verifyBackup(db, backupPath); err != nil {
		os.Remove(backupPath)
		return fmt.Errorf("backup verification failed: %w", err)
	}

	fmt.Println("=== Backup Completed Successfully ===")
	fmt.Printf("Backup location: %s\n", backupPath)
	return nil
}

func copyDatabaseFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source database: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	if err := destFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync backup file: %w", err)
	}

	return nil
}

func verifyBackup(originalDb *sql.DB, backupPath string) error {
	fmt.Println("Verifying backup integrity...")

	backupDb, err := sql.Open("libsql", "file:"+backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup database: %w", err)
	}
	defer backupDb.Close()

	tables := []string{"accounts", "transactions", "prices"}

	for _, table := range tables {
		var originalCount, backupCount int

		err := originalDb.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&originalCount)
		if err != nil {
			fmt.Printf("  Warning: Could not count %s in original (table may not exist): %v\n", table, err)
			continue
		}

		err = backupDb.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&backupCount)
		if err != nil {
			return fmt.Errorf("failed to count %s in backup: %w", table, err)
		}

		if originalCount != backupCount {
			return fmt.Errorf("count mismatch for %s: original=%d, backup=%d", table, originalCount, backupCount)
		}

		fmt.Printf("  ✓ %s: %d records verified\n", table, backupCount)
	}

	return nil
}
