package views

import (
	"fmt"
	"sort"
	"strings"

	"tracker/types"
	"tracker/utils"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AccountDetailView struct {
	table          table.Model
	account        types.Account
	portfolio      types.AnalyzedPortfolio
	transactions   []types.Transaction
	width          int
	height         int
	currencySymbol string
	exchangeRate   float64
	showDividends  bool
	styles         AccountDetailStyles
	focused        bool
}

type AccountDetailStyles struct {
	Container   lipgloss.Style
	Header      lipgloss.Style
	Title       lipgloss.Style
	Subtitle    lipgloss.Style
	InfoSection lipgloss.Style
	InfoLabel   lipgloss.Style
	InfoValue   lipgloss.Style
	Positive    lipgloss.Style
	Negative    lipgloss.Style
	Holdings    lipgloss.Style
	Tags        lipgloss.Style
}

func DefaultAccountDetailStyles() AccountDetailStyles {
	return AccountDetailStyles{
		Container: lipgloss.NewStyle().
			Padding(0, 1),
		Header: lipgloss.NewStyle().
			Background(lipgloss.Color("#312e81")).
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 2).
			MarginBottom(1),
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#22c55e")).
			Bold(true),
		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a5b4fc")),
		InfoSection: lipgloss.NewStyle().
			Padding(0, 2).
			MarginBottom(1),
		InfoLabel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280")),
		InfoValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#eaeaea")),
		Positive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#22c55e")),
		Negative: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ef4444")),
		Holdings: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#22d3ee")),
		Tags: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a5b4fc")).
			Italic(true),
	}
}

func NewAccountDetailView(account types.Account, portfolio types.AnalyzedPortfolio) AccountDetailView {
	v := AccountDetailView{
		account:        account,
		portfolio:      portfolio,
		transactions:   portfolio.Transactions,
		currencySymbol: "$",
		exchangeRate:   1.0,
		showDividends:  true,
		styles:         DefaultAccountDetailStyles(),
		focused:        true,
	}
	return v
}

func (v *AccountDetailView) SetSize(width, height int) {
	v.width = width
	v.height = height
	v.rebuildTable()
}

func (v *AccountDetailView) SetCurrency(symbol string, rate float64) {
	v.currencySymbol = symbol
	v.exchangeRate = rate
	v.rebuildTable()
}

func (v *AccountDetailView) ToggleDividends() bool {
	v.showDividends = !v.showDividends
	v.rebuildTable()
	return v.showDividends
}

func (v *AccountDetailView) SelectedTransaction() *types.Transaction {
	selected := v.table.SelectedRow()
	if len(selected) == 0 {
		return nil
	}

	cursor := v.table.Cursor()
	displayed := v.getDisplayedTransactions()
	if cursor >= 0 && cursor < len(displayed) {
		return &displayed[cursor]
	}
	return nil
}

func (v *AccountDetailView) Focus() {
	v.focused = true
	v.table.Focus()
}

func (v *AccountDetailView) Blur() {
	v.focused = false
	v.table.Blur()
}

func (v *AccountDetailView) Update(msg tea.Msg) (AccountDetailView, tea.Cmd) {
	var cmd tea.Cmd
	v.table, cmd = v.table.Update(msg)
	return *v, cmd
}

func (v *AccountDetailView) rebuildTable() {
	infoHeight := 8
	tableHeight := v.height - infoHeight - 2
	if tableHeight < 5 {
		tableHeight = 5
	}

	columns := v.buildColumns()
	rows := v.buildRows()

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(v.focused),
		table.WithHeight(tableHeight),
	)

	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#312e81")).
		Padding(0, 1)
	s.Cell = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#eaeaea")).
		Padding(0, 1)
	s.Selected = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#6366f1")).
		Padding(0, 1)
	t.SetStyles(s)

	v.table = t
}

func (v *AccountDetailView) buildColumns() []table.Column {
	w := v.width - 8
	return []table.Column{
		{Title: "Date", Width: w * 15 / 100},
		{Title: "Type", Width: w * 12 / 100},
		{Title: "Symbol", Width: w * 15 / 100},
		{Title: "Quantity", Width: w * 12 / 100},
		{Title: "Price", Width: w * 18 / 100},
		{Title: "Total", Width: w * 18 / 100},
	}
}

func (v *AccountDetailView) buildRows() []table.Row {
	displayed := v.getDisplayedTransactions()
	var rows []table.Row

	for _, tx := range displayed {
		total := int64(tx.Quantity) * int64(tx.Pps)
		rows = append(rows, table.Row{
			tx.Date.Format("2006-01-02"),
			string(tx.Type),
			tx.Symbol,
			fmt.Sprintf("%d", tx.Quantity),
			utils.ToCurrencyString(int64(tx.Pps), 2, v.currencySymbol, v.exchangeRate),
			utils.ToCurrencyString(total, 2, v.currencySymbol, v.exchangeRate),
		})
	}

	return rows
}

func (v *AccountDetailView) getDisplayedTransactions() []types.Transaction {
	sorted := make([]types.Transaction, len(v.transactions))
	copy(sorted, v.transactions)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Date.After(sorted[j].Date)
	})

	if v.showDividends {
		return sorted
	}

	var filtered []types.Transaction
	for _, tx := range sorted {
		if tx.Type != types.TransactionTypeDividend && tx.Type != types.TransactionTypeSplit {
			filtered = append(filtered, tx)
		}
	}
	return filtered
}

func (v AccountDetailView) View() string {
	header := v.renderHeader()
	info := v.renderInfo()
	tableView := v.table.View()

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		info,
		tableView,
	)
}

func (v AccountDetailView) renderHeader() string {
	title := v.styles.Title.Render(v.account.Name)
	subtitle := v.styles.Subtitle.Render(" │ " + v.account.Institution)

	return v.styles.Header.Width(v.width).Render(
		lipgloss.JoinHorizontal(lipgloss.Left, title, subtitle),
	)
}

func (v AccountDetailView) renderInfo() string {
	multiplier := v.exchangeRate

	col1 := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left,
			v.styles.InfoLabel.Render("Value: "),
			v.styles.Positive.Render(utils.ToCurrencyString(v.portfolio.Value, 0, v.currencySymbol, multiplier)),
		),
		lipgloss.JoinHorizontal(lipgloss.Left,
			v.styles.InfoLabel.Render("Invested: "),
			v.styles.InfoValue.Render(utils.ToCurrencyString(v.portfolio.TotalInvested, 0, v.currencySymbol, multiplier)),
		),
		lipgloss.JoinHorizontal(lipgloss.Left,
			v.styles.InfoLabel.Render("Dividends: "),
			v.styles.InfoValue.Render(utils.ToCurrencyString(v.portfolio.TotalDividends, 0, v.currencySymbol, multiplier)),
		),
	)

	gainStyle := v.styles.Positive
	if v.portfolio.Gain < 0 {
		gainStyle = v.styles.Negative
	}

	col2 := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left,
			v.styles.InfoLabel.Render("Gain: "),
			gainStyle.Render(utils.ToYieldString(v.portfolio.Gain)),
		),
		lipgloss.JoinHorizontal(lipgloss.Left,
			v.styles.InfoLabel.Render("Annualized: "),
			v.styles.InfoValue.Render(utils.ToYieldString(v.portfolio.AnnualizedYield)),
		),
		lipgloss.JoinHorizontal(lipgloss.Left,
			v.styles.InfoLabel.Render("Dietz: "),
			v.styles.InfoValue.Render(utils.ToYieldString(v.portfolio.ModifiedDietzYield)),
		),
	)

	var holdingsText string
	var symbols []string
	for symbol := range v.portfolio.SymbolsCount {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)
	for _, symbol := range symbols {
		count := v.portfolio.SymbolsCount[symbol]
		if count > 0 {
			holdingsText += fmt.Sprintf("%s(%d) ", symbol, count)
		}
	}
	if holdingsText == "" {
		holdingsText = "None"
	}

	col3 := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left,
			v.styles.InfoLabel.Render("Holdings: "),
			v.styles.Holdings.Render(holdingsText),
		),
		lipgloss.JoinHorizontal(lipgloss.Left,
			v.styles.InfoLabel.Render("Inception: "),
			v.styles.InfoValue.Render(v.portfolio.FirstTransaction.Date.Format("2006-01-02")),
		),
		lipgloss.JoinHorizontal(lipgloss.Left,
			v.styles.InfoLabel.Render("Tags: "),
			v.styles.Tags.Render(v.formatTags()),
		),
	)

	colWidth := (v.width - 8) / 3
	col1Styled := lipgloss.NewStyle().Width(colWidth).Render(col1)
	col2Styled := lipgloss.NewStyle().Width(colWidth).Render(col2)
	col3Styled := lipgloss.NewStyle().Width(colWidth).Render(col3)

	return v.styles.InfoSection.Render(
		lipgloss.JoinHorizontal(lipgloss.Top, col1Styled, col2Styled, col3Styled),
	)
}

func (v AccountDetailView) formatTags() string {
	if len(v.account.Tags) == 0 {
		return "-"
	}
	return strings.Join(v.account.Tags, ", ")
}
