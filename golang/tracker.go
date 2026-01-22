package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"tracker/btui"
	"tracker/market"
	"tracker/storage"
	"tracker/tui"
	"tracker/web"
)

func main() {
	tuiBackend := flag.String("tui", "tview", "TUI backend: 'tview' (classic) or 'bubble' (new bubbletea)")
	flag.Parse()

	args := flag.Args()

	if len(args) == 1 {
		switch args[0] {
		case "help":
			printHelp()
			return
		case "update":
			db, cleanup := storage.OpenDatabase()
			defer cleanup()
			market.UpdateMarketData(db)
			fmt.Println("Market data updated successfully")
			return
		case "server":
			web.StartServer()
			return
		case "backup":
			if err := runBackup(); err != nil {
				fmt.Fprintf(os.Stderr, "Backup failed: %v\n", err)
				os.Exit(1)
			}
			return
		default:
			fmt.Printf("Unknown command: %s\n", args[0])
			printHelp()
			return
		}
	} else if len(args) > 1 {
		fmt.Println("Too many arguments")
		printHelp()
		return
	}

	db, cleanup := storage.OpenDatabase()
	defer cleanup()

	switch strings.ToLower(*tuiBackend) {
	case "bubble", "bubbletea":
		btui.StartApp(db)
	case "tview", "classic":
		tui.StartApp(db)
	default:
		fmt.Printf("Unknown TUI backend: %s (use 'tview' or 'bubble')\n", *tuiBackend)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("Usage: tracker [options] [command]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  help     Show this help")
	fmt.Println("  update   Update market data")
	fmt.Println("  server   Start the web server")
	fmt.Println("  backup   Backup database to home directory")
	fmt.Println("  (none)   Start the portfolio tracker TUI")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -tui string")
	fmt.Println("        TUI backend: 'tview' (classic) or 'bubble' (new bubbletea)")
	fmt.Println("        Default: tview")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  tracker                  Start with classic tview TUI")
	fmt.Println("  tracker -tui=bubble      Start with new bubbletea TUI")
	fmt.Println("  tracker update           Update market data")
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
