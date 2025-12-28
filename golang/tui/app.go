package tui

import (
	"database/sql"
	"sync"
	"tracker/loaders"
	"tracker/market"
	"tracker/portfolio"
	"tracker/types"
	"tracker/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func fillAccountsTable(table *tview.Table, analysis *types.AnalysisData, theme Theme, tagFilter string, allPortfolioData types.AnalyzedPortfolio) {
	accounts := analysis.Accounts
	accountsData := analysis.AccountsData
	currency := analysis.ExchaneSign
	exchangeRate := analysis.ExchangeRate

	table.Clear()

	hs := tcell.StyleDefault.
		Background(theme.HeaderBg).
		Foreground(theme.HeaderFg).Bold(true)

	table.SetCell(0, 0, tview.NewTableCell("ID").SetStyle(hs).SetExpansion(1).SetAlign(tview.AlignLeft))
	table.SetCell(0, 1, tview.NewTableCell("Account Name").SetStyle(hs).SetExpansion(2).SetAlign(tview.AlignLeft))
	table.SetCell(0, 2, tview.NewTableCell("Total Invested").SetStyle(hs).SetExpansion(2).SetAlign(tview.AlignRight))
	table.SetCell(0, 3, tview.NewTableCell("Total Withdrawn").SetStyle(hs).SetExpansion(2).SetAlign(tview.AlignRight))
	table.SetCell(0, 4, tview.NewTableCell("Total Dividends").SetStyle(hs).SetExpansion(2).SetAlign(tview.AlignRight))
	table.SetCell(0, 5, tview.NewTableCell("Gain").SetStyle(hs).SetExpansion(1).SetAlign(tview.AlignRight))
	table.SetCell(0, 6, tview.NewTableCell("Annualized Yield").SetStyle(hs).SetExpansion(1).SetAlign(tview.AlignRight))
	table.SetCell(0, 7, tview.NewTableCell("Dietz Yield").SetStyle(hs).SetExpansion(1).SetAlign(tview.AlignRight))
	table.SetCell(0, 8, tview.NewTableCell("Value").SetStyle(hs).SetExpansion(2).SetAlign(tview.AlignRight))

	row := 1
	for _, ac := range *accounts {
		if tagFilter != "All" && !hasTag(ac.Tags, tagFilter) {
			continue
		}

		ts := tcell.StyleDefault
		table.SetCell(row, 0, tview.NewTableCell(ac.Id).SetExpansion(1).SetAlign(tview.AlignLeft))
		table.SetCell(row, 1, tview.NewTableCell(ac.Name).SetExpansion(2).SetAlign(tview.AlignLeft))

		acData, _ := accountsData[ac.Id]
		table.SetCell(row, 2, tview.NewTableCell(utils.ToCurrencyString(acData.TotalInvested, 0, currency, exchangeRate)).SetStyle(ts).SetExpansion(2).SetAlign(tview.AlignRight))
		table.SetCell(row, 3, tview.NewTableCell(utils.ToCurrencyString(acData.TotalWithdrawn, 0, currency, exchangeRate)).SetStyle(ts).SetExpansion(2).SetAlign(tview.AlignRight))
		table.SetCell(row, 4, tview.NewTableCell(utils.ToCurrencyString(acData.TotalDividends, 0, currency, exchangeRate)).SetStyle(ts).SetExpansion(2).SetAlign(tview.AlignRight))
		table.SetCell(row, 5, tview.NewTableCell(utils.ToYieldString(acData.Gain)).SetStyle(ts).SetExpansion(1).SetAlign(tview.AlignRight))
		table.SetCell(row, 6, tview.NewTableCell(utils.ToYieldString(acData.AnnualizedYield)).SetStyle(ts).SetExpansion(1).SetAlign(tview.AlignRight))
		table.SetCell(row, 7, tview.NewTableCell(utils.ToYieldString(acData.ModifiedDietzYield)).SetStyle(ts).SetExpansion(1).SetAlign(tview.AlignRight))
		table.SetCell(row, 8, tview.NewTableCell(utils.ToCurrencyString(acData.Value, 0, currency, exchangeRate)).SetStyle(ts).SetExpansion(2).SetAlign(tview.AlignRight))
		row++
	}

	ts := tcell.StyleDefault.Foreground(theme.Negative)
	table.SetCell(row, 0, tview.NewTableCell("").SetExpansion(1).SetAlign(tview.AlignLeft))
	table.SetCell(row, 1, tview.NewTableCell("All Portfolio").SetExpansion(2).SetAlign(tview.AlignLeft))
	table.SetCell(row, 2, tview.NewTableCell(utils.ToCurrencyString(allPortfolioData.TotalInvested, 0, currency, exchangeRate)).SetStyle(ts).SetExpansion(2).SetAlign(tview.AlignRight))
	table.SetCell(row, 3, tview.NewTableCell(utils.ToCurrencyString(allPortfolioData.TotalWithdrawn, 0, currency, exchangeRate)).SetStyle(ts).SetExpansion(2).SetAlign(tview.AlignRight))
	table.SetCell(row, 4, tview.NewTableCell(utils.ToCurrencyString(allPortfolioData.TotalDividends, 0, currency, exchangeRate)).SetStyle(ts).SetExpansion(2).SetAlign(tview.AlignRight))
	table.SetCell(row, 5, tview.NewTableCell(utils.ToYieldString(allPortfolioData.Gain)).SetStyle(ts).SetExpansion(1).SetAlign(tview.AlignRight))
	table.SetCell(row, 6, tview.NewTableCell(utils.ToYieldString(allPortfolioData.AnnualizedYield)).SetStyle(ts).SetExpansion(1).SetAlign(tview.AlignRight))
	table.SetCell(row, 7, tview.NewTableCell(utils.ToYieldString(allPortfolioData.ModifiedDietzYield)).SetStyle(ts).SetExpansion(1).SetAlign(tview.AlignRight))
	table.SetCell(row, 8, tview.NewTableCell(utils.ToCurrencyString(allPortfolioData.Value, 0, currency, exchangeRate)).SetStyle(ts).SetExpansion(2).SetAlign(tview.AlignRight))
}

func getFilteredAccountIds(accounts *[]types.Account, tagFilter string) []string {
	var ids []string
	for _, ac := range *accounts {
		if tagFilter == "All" || hasTag(ac.Tags, tagFilter) {
			ids = append(ids, ac.Id)
		}
	}
	return ids
}

func computeAllPortfolioData(db *sql.DB, analysis *types.AnalysisData, tagFilter string) types.AnalyzedPortfolio {
	filteredIds := getFilteredAccountIds(analysis.Accounts, tagFilter)
	data, _ := portfolio.LoadAndAnalyzeAccounts(db, filteredIds)
	return data
}

func hasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func collectUniqueTags(accounts *[]types.Account) []string {
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

func AccountsPage(db *sql.DB, analysis *types.AnalysisData, app *tview.Application, pages *tview.Pages) *tview.Flex {
	var selectedAccount int
	accounts := analysis.Accounts
	theme := GetTheme()
	currentTagFilter := "All"

	table := tview.NewTable().SetContent(nil)
	table.SetBorder(true).SetBorderColor(theme.Border)
	table.SetSelectable(true, false)
	table.SetSeparator('|').SetBorderPadding(2, 2, 3, 3)
	table.SetSelectedStyle(tcell.StyleDefault.Background(theme.SelectedBg).Foreground(theme.SelectedFg))

	allPortfolioData := computeAllPortfolioData(db, analysis, currentTagFilter)
	fillAccountsTable(table, analysis, theme, currentTagFilter, allPortfolioData)

	table.SetSelectionChangedFunc(func(row, column int) {
		if row == 0 {
			table.Select(1, column)
			selectedAccount = 0
		} else {
			selectedAccount = row - 1
		}
	})
	table.Select(1, 0).SetFixed(1, 2).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.Stop()
		}
	})

	table.SetSelectedFunc(func(row int, column int) {
		filteredAccounts := getFilteredAccounts(accounts, currentTagFilter)
		if selectedAccount >= len(filteredAccounts) {
			return
		}
		a := filteredAccounts[selectedAccount]
		if a.Id == "" {
			return
		}
		pages.AddAndSwitchToPage("account", SingleAccountPage(db, analysis, a, app, pages), true)
	})

	head := tview.NewTextView().SetText("All Accounts").SetTextAlign(tview.AlignCenter)
	focusItem := 0

	nisButton := tview.NewButton("View in NIS")
	nisButton.SetStyle(tcell.StyleDefault.Background(theme.ButtonBg).Foreground(theme.ButtonFg))
	nisButton.SetSelectedFunc(func() {
		val := loaders.CurrencyExchangeRate(db, "ILS")
		analysis.ExchaneSign = market.CurrencySymbolILS
		analysis.ExchangeRate = val
		allPortfolioData = computeAllPortfolioData(db, analysis, currentTagFilter)
		fillAccountsTable(table, analysis, theme, currentTagFilter, allPortfolioData)
		app.SetFocus(table)
		focusItem = 0
		app.Sync()
	})
	usdButton := tview.NewButton("View in USD")
	usdButton.SetStyle(tcell.StyleDefault.Background(theme.ButtonBg).Foreground(theme.ButtonFg))
	usdButton.SetSelectedFunc(func() {
		analysis.ExchaneSign = market.CurrencySymbolUSD
		analysis.ExchangeRate = 1.0
		allPortfolioData = computeAllPortfolioData(db, analysis, currentTagFilter)
		fillAccountsTable(table, analysis, theme, currentTagFilter, allPortfolioData)
		app.SetFocus(table)
		focusItem = 0
		app.Sync()
	})

	tagOptions := collectUniqueTags(accounts)
	tagDropdown := tview.NewDropDown().
		SetLabel("Tag: ").
		SetOptions(tagOptions, func(text string, index int) {
			currentTagFilter = text
			allPortfolioData = computeAllPortfolioData(db, analysis, currentTagFilter)
			fillAccountsTable(table, analysis, theme, currentTagFilter, allPortfolioData)
			app.Sync()
		}).
		SetCurrentOption(0)
	tagDropdown.SetFieldBackgroundColor(theme.ButtonBg)
	tagDropdown.SetFieldTextColor(theme.ButtonFg)

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	buttonFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	buttonFlex.AddItem(tview.NewTextView().SetText(" "), 0, 1, false)
	buttonFlex.AddItem(nisButton, 0, 1, false)
	buttonFlex.AddItem(tview.NewTextView().SetText(" "), 0, 2, false)
	buttonFlex.AddItem(usdButton, 0, 1, false)
	buttonFlex.AddItem(tview.NewTextView().SetText(" "), 0, 2, false)
	buttonFlex.AddItem(tagDropdown, 0, 2, false)
	buttonFlex.AddItem(tview.NewTextView().SetText(" "), 0, 1, false)

	flex.AddItem(head, 1, 1, false)
	flex.AddItem(buttonFlex, 1, 0, false)
	flex.AddItem(table, 0, 1, true)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			if focusItem == 0 {
				app.SetFocus(nisButton)
				focusItem = 1
			} else if focusItem == 1 {
				app.SetFocus(usdButton)
				focusItem = 2
			} else if focusItem == 2 {
				app.SetFocus(tagDropdown)
				focusItem = 3
			} else {
				app.SetFocus(table)
				focusItem = 0
			}
			return nil
		}

		return event
	})

	return flex
}

func getFilteredAccounts(accounts *[]types.Account, tagFilter string) []types.Account {
	var filtered []types.Account
	for _, ac := range *accounts {
		if tagFilter == "All" || hasTag(ac.Tags, tagFilter) {
			filtered = append(filtered, ac)
		}
	}
	return filtered
}

func StartApp(db *sql.DB) {
	theme := GetTheme()

	accounts, _ := loaders.UserAccounts(db)

	accountsData := make(map[string]types.AnalyzedPortfolio, len(*accounts))
	var wg sync.WaitGroup
	for i := range *accounts {
		wg.Go(func() {
			ac := (*accounts)[i]
			data, _ := portfolio.LoadAndAnalyze(db, ac)
			accountsData[ac.Id] = data
		})
	}
	wg.Wait()

	analysis := types.NewAnalysisData(accounts, accountsData)

	app := tview.NewApplication()
	pages := tview.NewPages()
	pages.SetBackgroundColor(theme.Background)
	pages.AddPage("Accounts", AccountsPage(db, &analysis, app, pages), true, true)

	app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		screen.Fill(' ', tcell.StyleDefault.Background(theme.Background).Foreground(theme.Foreground))
		return false
	})

	if err := app.SetRoot(pages, true).SetFocus(pages).Run(); err != nil {
		panic(err)
	}

}
