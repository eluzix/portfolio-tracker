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

func StartApp(db *sql.DB) {
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

	app := tview.NewApplication()
	table := tview.NewTable().SetContent(nil)
	table.SetBorder(true).SetBorderStyle(tcell.StyleDefault.Background(tcell.ColorRed))
	table.SetSelectable(true, false)

	table.SetCell(0, 0, tview.NewTableCell("ID").SetBackgroundColor(tcell.ColorLavender))
	table.SetCell(0, 1, tview.NewTableCell("Account Name"))
	table.SetCell(0, 2, tview.NewTableCell("Total Invested"))
	table.SetCell(0, 3, tview.NewTableCell("Total Withdrawn"))
	table.SetCell(0, 4, tview.NewTableCell("Total Dividends"))
	table.SetCell(0, 5, tview.NewTableCell("Value"))
	for i, ac := range *accounts {
		table.SetCell(i, 0, tview.NewTableCell(ac.Id))
		table.SetCell(i, 1, tview.NewTableCell(ac.Name))

		acData, _ := accountsData[ac.Id]
		table.SetCell(i, 2, tview.NewTableCell(utils.ToCurrencyString(acData.TotalInvested)))
		table.SetCell(i, 3, tview.NewTableCell(utils.ToCurrencyString(acData.TotalWithdrawn)))
		table.SetCell(i, 4, tview.NewTableCell(utils.ToCurrencyString(acData.TotalDividends)))
		table.SetCell(i, 5, tview.NewTableCell(utils.ToCurrencyString(acData.Value)))
	}

	table.Select(0, 0).SetFixed(1, 1).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.Stop()
		}
		// if key == tcell.KeyEnter {
		// 	table.SetSelectable(true, true)
		// }
	})
	// .SetSelectedFunc(func(row int, column int) {
	// 	table.GetCell(row, column).SetTextColor(tcell.ColorRed)
	// 	table.SetSelectable(false, false)
	// })

	if err := app.SetRoot(table, true).SetFocus(table).Run(); err != nil {
		panic(err)
	}

}
