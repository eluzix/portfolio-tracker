package tui

import (
	"database/sql"
	"sync"
	"tracker/loaders"
	"tracker/portfolio"
	"tracker/types"
	"tracker/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func SingleAccountPage(account types.Account, portfolio types.AnalyzedPortfolio, app *tview.Application, pages *tview.Pages) tview.Primitive {
	newPrimitive := func(text string) tview.Primitive {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(text)
	}

	grid := tview.NewGrid().
		SetRows(100, 20).
		SetColumns(100, 20).
		SetBorders(true).
		AddItem(newPrimitive("head"), 1, 0, 1, 1, 0, 0, true).
		AddItem(newPrimitive("body"), 0, 0, 1, 1, 0, 0, false)
	return grid
}

func AccountsPage(accounts *[]types.Account, accountsData map[string]types.AnalyzedPortfolio, app *tview.Application, pages *tview.Pages) *tview.Table {
	var selectedAccount int

	table := tview.NewTable().SetContent(nil)
	table.SetBorder(true).SetBorderColor(tcell.ColorGreenYellow)
	table.SetSelectable(true, false)
	table.SetSeparator('|').SetBorderPadding(0, 1, 1, 1)

	hs := tcell.StyleDefault.
		Background(tcell.ColorOrangeRed).
		Foreground(tcell.ColorGreenYellow).Bold(true)

	table.SetCell(0, 0, tview.NewTableCell("ID").SetStyle(hs))
	table.SetCell(0, 1, tview.NewTableCell("Account Name").SetStyle(hs))
	table.SetCell(0, 2, tview.NewTableCell("Total Invested").SetStyle(hs))
	table.SetCell(0, 3, tview.NewTableCell("Total Withdrawn").SetStyle(hs))
	table.SetCell(0, 4, tview.NewTableCell("Total Dividends").SetStyle(hs))
	table.SetCell(0, 5, tview.NewTableCell("Gain").SetStyle(hs))
	table.SetCell(0, 6, tview.NewTableCell("Annualized Yield").SetStyle(hs))
	table.SetCell(0, 7, tview.NewTableCell("Dietz Yield").SetStyle(hs))
	table.SetCell(0, 8, tview.NewTableCell("Value").SetStyle(hs))
	for i, ac := range *accounts {
		ts := tcell.StyleDefault
		if i == len(*accounts) {
			ts = ts.Foreground(tcell.ColorRed)
		}

		table.SetCell(i+1, 0, tview.NewTableCell(ac.Id))
		table.SetCell(i+1, 1, tview.NewTableCell(ac.Name))

		acData, _ := accountsData[ac.Id]
		table.SetCell(i+1, 2, tview.NewTableCell(utils.ToCurrencyString(acData.TotalInvested)).SetStyle(ts))
		table.SetCell(i+1, 3, tview.NewTableCell(utils.ToCurrencyString(acData.TotalWithdrawn)).SetStyle(ts))
		table.SetCell(i+1, 4, tview.NewTableCell(utils.ToCurrencyString(acData.TotalDividends)).SetStyle(ts))
		table.SetCell(i+1, 5, tview.NewTableCell(utils.ToYieldString(acData.Gain)).SetStyle(ts))
		table.SetCell(i+1, 6, tview.NewTableCell(utils.ToYieldString(acData.AnnualizedYield)).SetStyle(ts))
		table.SetCell(i+1, 7, tview.NewTableCell(utils.ToYieldString(acData.ModifiedDietzYield)).SetStyle(ts))
		table.SetCell(i+1, 8, tview.NewTableCell(utils.ToCurrencyString(acData.Value)).SetStyle(ts))
	}

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
		pages.AddAndSwitchToPage("account", SingleAccountPage(a, accountsData[a.Id], app, pages), false)
	})
	return table
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

	app := tview.NewApplication()
	pages := tview.NewPages()
	pages.AddPage("Accounts", AccountsPage(accounts, accountsData, app, pages), true, true)

	if err := app.SetRoot(pages, true).SetFocus(pages).Run(); err != nil {
		panic(err)
	}

}
