package views

import (
	"fmt"
	"slices"

	"tracker/types"
	"tracker/utils"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AccountsView struct {
	table          table.Model
	accounts       *[]types.Account
	accountsData   map[string]types.AnalyzedPortfolio
	allPortfolio   types.AnalyzedPortfolio
	width          int
	height         int
	currencySymbol string
	exchangeRate   float64
	tagFilter      string
	tags           []string
	tagIndex       int
	styles         AccountsStyles
	focused        bool
}

type AccountsStyles struct {
	Container  lipgloss.Style
	InfoBar    lipgloss.Style
	InfoLabel  lipgloss.Style
	InfoValue  lipgloss.Style
	Positive   lipgloss.Style
	Negative   lipgloss.Style
	TotalRow   lipgloss.Style
}

func DefaultAccountsStyles() AccountsStyles {
	return AccountsStyles{
		Container: lipgloss.NewStyle().
			Padding(1, 2),
		InfoBar: lipgloss.NewStyle().
			Padding(0, 2).
			MarginBottom(1),
		InfoLabel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#737aa2")),
		InfoValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7dcfff")).
			Bold(true),
		Positive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9ece6a")),
		Negative: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f7768e")),
		TotalRow: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e0af68")).
			Bold(true),
	}
}

func NewAccountsView(accounts *[]types.Account, accountsData map[string]types.AnalyzedPortfolio, allPortfolio types.AnalyzedPortfolio) AccountsView {
	v := AccountsView{
		accounts:       accounts,
		accountsData:   accountsData,
		allPortfolio:   allPortfolio,
		currencySymbol: "$",
		exchangeRate:   1.0,
		tagFilter:      "All",
		tags:           collectTags(accounts),
		styles:         DefaultAccountsStyles(),
		focused:        true,
	}
	return v
}

func (v *AccountsView) SetSize(width, height int) {
	v.width = width
	v.height = height
	v.rebuildTable()
}

func (v *AccountsView) SetCurrency(symbol string, rate float64) {
	v.currencySymbol = symbol
	v.exchangeRate = rate
	v.rebuildTable()
}

func (v *AccountsView) SetTagFilter(tag string) {
	v.tagFilter = tag
	v.rebuildTable()
}

func (v *AccountsView) CycleTag() string {
	v.tagIndex = (v.tagIndex + 1) % len(v.tags)
	v.tagFilter = v.tags[v.tagIndex]
	v.rebuildTable()
	return v.tagFilter
}

func (v *AccountsView) SetAllPortfolio(p types.AnalyzedPortfolio) {
	v.allPortfolio = p
	v.rebuildTable()
}

func (v *AccountsView) SelectedAccount() *types.Account {
	selected := v.table.SelectedRow()
	if len(selected) == 0 {
		return nil
	}
	id := selected[0]
	if id == "" {
		return nil
	}
	for _, ac := range *v.accounts {
		if ac.Id == id {
			return &ac
		}
	}
	return nil
}

func (v *AccountsView) Focus() {
	v.focused = true
	v.table.Focus()
}

func (v *AccountsView) Blur() {
	v.focused = false
	v.table.Blur()
}

func (v *AccountsView) Update(msg tea.Msg) (AccountsView, tea.Cmd) {
	var cmd tea.Cmd
	v.table, cmd = v.table.Update(msg)
	return *v, cmd
}

func (v *AccountsView) rebuildTable() {
	tableHeight := v.height - 4
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
		Foreground(lipgloss.Color("#1a1b26")).
		Background(lipgloss.Color("#bb9af7")).
		Padding(0, 1)
	s.Cell = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e0e6f0")).
		Padding(0, 1)
	s.Selected = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#1a1b26")).
		Background(lipgloss.Color("#d4a5c8")).
		Padding(0, 1)
	t.SetStyles(s)

	v.table = t
}

func (v *AccountsView) buildColumns() []table.Column {
	w := v.width - 6
	return []table.Column{
		{Title: "ID", Width: w * 8 / 100},
		{Title: "Account Name", Width: w * 16 / 100},
		{Title: "Value", Width: w * 12 / 100},
		{Title: "Invested", Width: w * 11 / 100},
		{Title: "Withdrawn", Width: w * 11 / 100},
		{Title: "Dividends", Width: w * 10 / 100},
		{Title: "Gain", Width: w * 10 / 100},
		{Title: "Annual", Width: w * 10 / 100},
		{Title: "Dietz", Width: w * 10 / 100},
	}
}

func (v *AccountsView) buildRows() []table.Row {
	var rows []table.Row

	for _, ac := range *v.accounts {
		if v.tagFilter != "All" && !hasTag(ac.Tags, v.tagFilter) {
			continue
		}

		data := v.accountsData[ac.Id]
		rows = append(rows, table.Row{
			ac.Id,
			ac.Name,
			utils.ToCurrencyString(data.Value, 0, v.currencySymbol, v.exchangeRate),
			utils.ToCurrencyString(data.TotalInvested, 0, v.currencySymbol, v.exchangeRate),
			utils.ToCurrencyString(data.TotalWithdrawn, 0, v.currencySymbol, v.exchangeRate),
			utils.ToCurrencyString(data.TotalDividends, 0, v.currencySymbol, v.exchangeRate),
			utils.ToYieldString(data.Gain),
			utils.ToYieldString(data.AnnualizedYield),
			utils.ToYieldString(data.ModifiedDietzYield),
		})
	}

	rows = append(rows, table.Row{
		"",
		"── All Portfolio ──",
		utils.ToCurrencyString(v.allPortfolio.Value, 0, v.currencySymbol, v.exchangeRate),
		utils.ToCurrencyString(v.allPortfolio.TotalInvested, 0, v.currencySymbol, v.exchangeRate),
		utils.ToCurrencyString(v.allPortfolio.TotalWithdrawn, 0, v.currencySymbol, v.exchangeRate),
		utils.ToCurrencyString(v.allPortfolio.TotalDividends, 0, v.currencySymbol, v.exchangeRate),
		utils.ToYieldString(v.allPortfolio.Gain),
		utils.ToYieldString(v.allPortfolio.AnnualizedYield),
		utils.ToYieldString(v.allPortfolio.ModifiedDietzYield),
	})

	return rows
}

func (v AccountsView) View() string {
	infoBar := lipgloss.JoinHorizontal(lipgloss.Left,
		v.styles.InfoLabel.Render("Currency: "),
		v.styles.InfoValue.Render(v.currencySymbol),
		v.styles.InfoLabel.Render("  │  Tag: "),
		v.styles.InfoValue.Render(v.tagFilter),
		v.styles.InfoLabel.Render(fmt.Sprintf("  │  Accounts: %d", v.filteredCount())),
	)

	tableView := v.table.View()

	return lipgloss.JoinVertical(lipgloss.Left,
		v.styles.InfoBar.Render(infoBar),
		tableView,
	)
}

func (v *AccountsView) filteredCount() int {
	if v.tagFilter == "All" {
		return len(*v.accounts)
	}
	count := 0
	for _, ac := range *v.accounts {
		if hasTag(ac.Tags, v.tagFilter) {
			count++
		}
	}
	return count
}

func collectTags(accounts *[]types.Account) []string {
	if accounts == nil {
		return []string{"All"}
	}
	tagSet := make(map[string]bool)
	for _, ac := range *accounts {
		for _, tag := range ac.Tags {
			tagSet[tag] = true
		}
	}
	tags := []string{"All"}
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	return tags
}

func hasTag(tags []string, tag string) bool {
	return slices.Contains(tags, tag)
}
