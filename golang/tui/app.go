package tui

import (
	"database/sql"
	"fmt"
	"os"
	"tracker/config"

	tea "github.com/charmbracelet/bubbletea"
)

func StartApp(db *sql.DB, cfg config.AppConfig) {
	p := tea.NewProgram(
		NewModel(db, cfg),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
