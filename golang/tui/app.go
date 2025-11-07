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

func fillAccountsTable(table *tview.Table, analysis *types.AnalysisData) {
	accounts := analysis.Accounts
	accountsData := analysis.AccountsData
	currency := analysis.ExchaneSign
	exchangeRate := analysis.ExchangeRate

	for i, ac := range *accounts {
		ts := tcell.StyleDefault
		if i == len(*accounts) {
			ts = ts.Foreground(tcell.ColorRed)
		}

		table.SetCell(i+1, 0, tview.NewTableCell(ac.Id).SetExpansion(1).SetAlign(tview.AlignLeft))
		table.SetCell(i+1, 1, tview.NewTableCell(ac.Name).SetExpansion(2).SetAlign(tview.AlignLeft))

		acData, _ := accountsData[ac.Id]
		table.SetCell(i+1, 2, tview.NewTableCell(utils.ToCurrencyString(acData.TotalInvested, 0, currency, exchangeRate)).SetStyle(ts).SetExpansion(2).SetAlign(tview.AlignRight))
		table.SetCell(i+1, 3, tview.NewTableCell(utils.ToCurrencyString(acData.TotalWithdrawn, 0, currency, exchangeRate)).SetStyle(ts).SetExpansion(2).SetAlign(tview.AlignRight))
		table.SetCell(i+1, 4, tview.NewTableCell(utils.ToCurrencyString(acData.TotalDividends, 0, currency, exchangeRate)).SetStyle(ts).SetExpansion(2).SetAlign(tview.AlignRight))
		table.SetCell(i+1, 5, tview.NewTableCell(utils.ToYieldString(acData.Gain)).SetStyle(ts).SetExpansion(1).SetAlign(tview.AlignRight))
		table.SetCell(i+1, 6, tview.NewTableCell(utils.ToYieldString(acData.AnnualizedYield)).SetStyle(ts).SetExpansion(1).SetAlign(tview.AlignRight))
		table.SetCell(i+1, 7, tview.NewTableCell(utils.ToYieldString(acData.ModifiedDietzYield)).SetStyle(ts).SetExpansion(1).SetAlign(tview.AlignRight))
		table.SetCell(i+1, 8, tview.NewTableCell(utils.ToCurrencyString(acData.Value, 0, currency, exchangeRate)).SetStyle(ts).SetExpansion(2).SetAlign(tview.AlignRight))
	}
}

func AccountsPage(db *sql.DB, analysis *types.AnalysisData, app *tview.Application, pages *tview.Pages) *tview.Flex {
	var selectedAccount int
	accounts := analysis.Accounts

	table := tview.NewTable().SetContent(nil)
	table.SetBorder(true).SetBorderColor(tcell.ColorGreenYellow)
	table.SetSelectable(true, false)
	table.SetSeparator('|').SetBorderPadding(2, 2, 3, 3)

	hs := tcell.StyleDefault.
		Background(tcell.ColorOrangeRed).
		Foreground(tcell.ColorGreenYellow).Bold(true)

	table.SetCell(0, 0, tview.NewTableCell("ID").SetStyle(hs).SetExpansion(1).SetAlign(tview.AlignLeft))
	table.SetCell(0, 1, tview.NewTableCell("Account Name").SetStyle(hs).SetExpansion(2).SetAlign(tview.AlignLeft))
	table.SetCell(0, 2, tview.NewTableCell("Total Invested").SetStyle(hs).SetExpansion(2).SetAlign(tview.AlignRight))
	table.SetCell(0, 3, tview.NewTableCell("Total Withdrawn").SetStyle(hs).SetExpansion(2).SetAlign(tview.AlignRight))
	table.SetCell(0, 4, tview.NewTableCell("Total Dividends").SetStyle(hs).SetExpansion(2).SetAlign(tview.AlignRight))
	table.SetCell(0, 5, tview.NewTableCell("Gain").SetStyle(hs).SetExpansion(1).SetAlign(tview.AlignRight))
	table.SetCell(0, 6, tview.NewTableCell("Annualized Yield").SetStyle(hs).SetExpansion(1).SetAlign(tview.AlignRight))
	table.SetCell(0, 7, tview.NewTableCell("Dietz Yield").SetStyle(hs).SetExpansion(1).SetAlign(tview.AlignRight))
	table.SetCell(0, 8, tview.NewTableCell("Value").SetStyle(hs).SetExpansion(2).SetAlign(tview.AlignRight))

	fillAccountsTable(table, analysis)

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
		a := (*accounts)[selectedAccount]
		if a.Id == "" {
			return
		}
		pages.AddAndSwitchToPage("account", SingleAccountPage(db, analysis, a, app, pages), true)
	})

	head := tview.NewTextView().SetText("All Accounts").SetTextAlign(tview.AlignCenter)
	focusItem := 0

	nisButton := tview.NewButton("View in NIS").SetSelectedFunc(func() {
		val := loaders.CurrencyExchangeRate(db, "ILS")
		analysis.ExchaneSign = market.CurrencySymbolILS
		analysis.ExchangeRate = val
		fillAccountsTable(table, analysis)
		app.SetFocus(table)
		focusItem = 0

		go func() {
			app.Draw()
		}()
	})
	usdButton := tview.NewButton("View in USD").SetSelectedFunc(func() {
		analysis.ExchaneSign = market.CurrencySymbolUSD
		analysis.ExchangeRate = 1.0
		fillAccountsTable(table, analysis)
		app.SetFocus(table)
		focusItem = 0

		go func() {
			app.Draw()
		}()
	})

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	buttonFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	buttonFlex.AddItem(tview.NewTextView().SetText(" "), 0, 1, false)
	buttonFlex.AddItem(nisButton, 0, 1, false)
	buttonFlex.AddItem(tview.NewTextView().SetText(" "), 0, 2, false)
	buttonFlex.AddItem(usdButton, 0, 1, false)
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

func StartApp(db *sql.DB) {
	accounts, _ := loaders.UserAccounts(db)
	ac := types.Account{
		Id:   "",
		Name: "All Portfolio",
	}
	(*accounts) = append((*accounts), ac)

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
	pages.AddPage("Accounts", AccountsPage(db, &analysis, app, pages), true, true)

	if err := app.SetRoot(pages, true).SetFocus(pages).Run(); err != nil {
		panic(err)
	}

}
