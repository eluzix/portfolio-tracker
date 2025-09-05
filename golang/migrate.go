package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"tracker/storage"
	"tracker/types"
)

// Change this to choose which migration to run
const (
	MIGRATE_TRANSACTIONS = "transactions"
	MIGRATE_ACCOUNTS     = "accounts"
	MIGRATE_BOTH         = "both"
	COUNT_TRANSACTIONS   = "count_t"
)

// Set this to control which migration runs
// var MIGRATION_TYPE = COUNT_TRANSACTIONS
var MIGRATION_TYPE = MIGRATE_TRANSACTIONS

func main() {
	db, cleanup := storage.OpenDatabase()
	defer cleanup()

	switch MIGRATION_TYPE {
	case MIGRATE_TRANSACTIONS:
		migrateTransactions(db)
	case MIGRATE_ACCOUNTS:
		migrateAccounts(db)
	case MIGRATE_BOTH:
		migrateAccounts(db)
		migrateTransactions(db)
	case COUNT_TRANSACTIONS:
		countTransactions(db)
	default:
		fmt.Fprintf(os.Stderr, "Unknown migration type: %s\n", MIGRATION_TYPE)
		os.Exit(1)
	}
}

func countTransactions(db *sql.DB) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM transactions").Scan(&count)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to count transactions: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Total transactions in database: %d\n", count)
}

func migrateTransactions(db *sql.DB) {
	fmt.Println("=== Migrating Transactions ===")

	// Create transactions table if it doesn't exist
	// createTableSQL := `
	// CREATE TABLE IF NOT EXISTS transactions (
	// 	id TEXT PRIMARY KEY,
	// 	account_id TEXT NOT NULL,
	// 	symbol TEXT NOT NULL,
	// 	date TEXT NOT NULL,
	// 	transaction_type TEXT NOT NULL,
	// 	quantity INTEGER NOT NULL,
	// 	pps INTEGER NOT NULL
	// )`
	//
	// _, err := db.Exec(createTableSQL)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "failed to create transactions table: %v\n", err)
	// 	os.Exit(1)
	// }
	// fmt.Println("Transactions table created or already exists")

	// Open and read transactions.jsonl file
	file, err := os.Open("transactions.jsonl")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open transactions.jsonl: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Prepare insert statement

	tx, err := db.Begin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start transaction: %v\n", err)
		os.Exit(1)
	}
	defer tx.Rollback()

	// _, err = tx.Exec("DEL transactions")
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "failed to truncate transactions table: %v\n", err)
	// 	os.Exit(1)
	// }

	insertSQL := `
	INSERT OR REPLACE INTO transactions (id, account_id, symbol, date, transaction_type, quantity, pps)
	VALUES (?, ?, ?, ?, ?, ?, ?)`

	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to prepare insert statement: %v\n", err)
		os.Exit(1)
	}
	defer stmt.Close()

	// Read file line by line and insert transactions
	scanner := bufio.NewScanner(file)
	transactionCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		var transaction types.Transaction
		if err := json.Unmarshal([]byte(line), &transaction); err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse transaction JSON: %v\nLine: %s\n", err, line)
			continue
		}

		_, err := stmt.Exec(
			transaction.Id,
			transaction.AccountId,
			transaction.Symbol,
			transaction.Date,
			transaction.Type,
			transaction.Quantity,
			transaction.Pps,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to insert transaction %s: %v\n", transaction.Id, err)
			continue
		}

		transactionCount++
		if transactionCount%5 == 0 {
			fmt.Printf("Inserted %d transactions...\n", transactionCount)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
		os.Exit(1)
	}

	if err = tx.Commit(); err != nil {
		fmt.Fprintf(os.Stderr, "error commiting transaction: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully inserted %d transactions into the database\n", transactionCount)

	// Verify by counting total records
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM transactions").Scan(&count)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to count transactions: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Total transactions in database: %d\n", count)
}

func migrateAccounts(db *sql.DB) {
	fmt.Println("=== Migrating Accounts ===")

	// Create accounts table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS accounts (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		owner TEXT NOT NULL,
		institution TEXT NOT NULL,
		institution_id TEXT NOT NULL,
		description TEXT,
		tags TEXT,
		created_at TEXT,
		updated_at TEXT
	)`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create accounts table: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Accounts table created or already exists")

	// Open and read accounts.jsonl file
	file, err := os.Open("accounts.jsonl")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open accounts.jsonl: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Prepare insert statement
	insertSQL := `
	INSERT OR REPLACE INTO accounts (id, name, owner, institution, institution_id, description, tags, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to prepare insert statement: %v\n", err)
		os.Exit(1)
	}
	defer stmt.Close()

	// Read file line by line and insert accounts
	scanner := bufio.NewScanner(file)
	accountCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		var account types.Account
		if err := json.Unmarshal([]byte(line), &account); err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse account JSON: %v\nLine: %s\n", err, line)
			continue
		}

		// Convert tags slice to comma-separated string for storage
		var tagsStr string
		if len(account.Tags) > 0 {
			tagsStr = strings.Join(account.Tags, ",")
		}

		var description, createdAt, updatedAt interface{}
		if account.Description != nil {
			description = *account.Description
		}
		if account.CreatedAt != nil {
			createdAt = *account.CreatedAt
		}
		if account.UpdatedAt != nil {
			updatedAt = *account.UpdatedAt
		}

		_, err := stmt.Exec(
			account.Id,
			account.Name,
			account.Owner,
			account.Institution,
			account.InstitutionId,
			description,
			tagsStr,
			createdAt,
			updatedAt,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to insert account %s: %v\n", account.Id, err)
			continue
		}

		accountCount++
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully inserted %d accounts into the database\n", accountCount)

	// Verify by counting total records
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM accounts").Scan(&count)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to count accounts: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Total accounts in database: %d\n", count)
}
