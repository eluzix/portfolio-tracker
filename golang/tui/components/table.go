package components

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Table struct {
	table  table.Model
	width  int
	height int
	styles TableStyles
}

type TableStyles struct {
	Container lipgloss.Style
	Header    lipgloss.Style
	Cell      lipgloss.Style
	Selected  lipgloss.Style
	Border    lipgloss.Style
}

func DefaultTableStyles() TableStyles {
	return TableStyles{
		Container: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#475569")),
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#312e81")).
			Padding(0, 1),
		Cell: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#eaeaea")).
			Padding(0, 1),
		Selected: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#6366f1")).
			Padding(0, 1),
		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#475569")),
	}
}

func NewTable(columns []table.Column, rows []table.Row, width, height int) Table {
	styles := DefaultTableStyles()

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height-2),
	)

	s := table.DefaultStyles()
	s.Header = styles.Header
	s.Cell = styles.Cell
	s.Selected = styles.Selected
	t.SetStyles(s)

	return Table{
		table:  t,
		width:  width,
		height: height,
		styles: styles,
	}
}

func (t *Table) SetSize(width, height int) {
	t.width = width
	t.height = height
	t.table.SetHeight(height - 2)
}

func (t *Table) SetRows(rows []table.Row) {
	t.table.SetRows(rows)
}

func (t *Table) SetColumns(columns []table.Column) {
	t.table.SetColumns(columns)
}

func (t *Table) SelectedRow() table.Row {
	return t.table.SelectedRow()
}

func (t *Table) Cursor() int {
	return t.table.Cursor()
}

func (t *Table) SetCursor(cursor int) {
	t.table.SetCursor(cursor)
}

func (t *Table) Focus() {
	t.table.Focus()
}

func (t *Table) Blur() {
	t.table.Blur()
}

func (t *Table) Focused() bool {
	return t.table.Focused()
}

func (t *Table) Update(msg tea.Msg) (Table, tea.Cmd) {
	var cmd tea.Cmd
	t.table, cmd = t.table.Update(msg)
	return *t, cmd
}

func (t Table) View() string {
	tableView := t.table.View()
	return t.styles.Container.
		Width(t.width).
		Render(tableView)
}

func CalculateColumnWidths(totalWidth int, proportions []int) []int {
	padding := 4
	available := totalWidth - padding
	total := 0
	for _, p := range proportions {
		total += p
	}

	widths := make([]int, len(proportions))
	for i, p := range proportions {
		widths[i] = (available * p) / total
	}
	return widths
}
