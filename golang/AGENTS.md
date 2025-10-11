# Agent Instructions for Portfolio Tracker

## Build/Test Commands
- **Build**: `make build` or `go build -o tracker tracker.go`
- **Run**: `make run` or `go run tracker.go`
- **Test all**: `make test` or `go test ./...`
- **Test single package**: `go test ./portfolio/` (or specific package)
- **Test single function**: `go test -run TestFunctionName ./package/`
- **Clean**: `make clean`

## Architecture
- **Go portfolio tracker** with TUI using tview library
- **Database**: LibSQL/SQLite with cloud replication support (Turso)
- **Core packages**: `types/` (domain), `portfolio/` (analysis), `market/` (data fetching), `storage/` (DB), `tui/` (interface)
- **Data sources**: MarketStack API for prices, JSONL files for migrations
- **Pattern**: Repository pattern via loaders, interface-based design for market data

## Code Style
- **Imports**: Standard lib → Third-party → Local (`tracker/package`)
- **Naming**: PascalCase exports, camelCase private, lowercase packages
- **Errors**: Explicit checking, `fmt.Errorf` wrapping, custom types for domain errors
- **Testing**: Table-driven tests, fuzzing, property-based validation in `_test.go` files
- **Concurrency**: `sync.WaitGroup` for coordination, goroutines for parallel data fetching
- **DB**: Defer cleanup, transaction patterns, prepared statements
- **Comments**: NEVER add code comments when generating code unless explicitly requested
