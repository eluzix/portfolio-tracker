package tui

import (
	"fmt"
	"sort"
	"tracker/types"
	"tracker/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TransactionsTable(portfolio types.AnalyzedPortfolio, pages *tview.Pages) *tview.Table {
	transactionsTable := tview.NewTable().SetContent(nil)
	transactionsTable.SetBorder(true).SetBorderColor(tcell.ColorGreenYellow)
	transactionsTable.SetSelectable(true, false)
	transactionsTable.SetSeparator('|').SetBorderPadding(0, 1, 1, 1)

	headerStyle := tcell.StyleDefault.
		Background(tcell.ColorOrangeRed).
		Foreground(tcell.ColorGreenYellow).Bold(true)

	transactionsTable.SetCell(0, 0, tview.NewTableCell("Date").SetStyle(headerStyle))
	transactionsTable.SetCell(0, 1, tview.NewTableCell("Type").SetStyle(headerStyle))
	transactionsTable.SetCell(0, 2, tview.NewTableCell("Symbol").SetStyle(headerStyle))
	transactionsTable.SetCell(0, 3, tview.NewTableCell("Quantity").SetStyle(headerStyle))
	transactionsTable.SetCell(0, 4, tview.NewTableCell("Price").SetStyle(headerStyle))
	transactionsTable.SetCell(0, 5, tview.NewTableCell("Total").SetStyle(headerStyle))

	sortedTransactions := make([]types.Transaction, len(portfolio.Transactions))
	copy(sortedTransactions, portfolio.Transactions)
	sort.Slice(sortedTransactions, func(i, j int) bool {
		return sortedTransactions[i].Date.After(sortedTransactions[j].Date)
	})

	for i, tx := range sortedTransactions {
		row := i + 1
		transactionsTable.SetCell(row, 0, tview.NewTableCell(tx.Date.Format("2006-01-02")))
		transactionsTable.SetCell(row, 1, tview.NewTableCell(string(tx.Type)))
		transactionsTable.SetCell(row, 2, tview.NewTableCell(tx.Symbol))
		transactionsTable.SetCell(row, 3, tview.NewTableCell(fmt.Sprintf("%d", tx.Quantity)))
		transactionsTable.SetCell(row, 4, tview.NewTableCell(utils.ToCurrencyString(int64(tx.Pps), 2)))

		total := int64(tx.Quantity) * int64(tx.Pps)
		transactionsTable.SetCell(row, 5, tview.NewTableCell(utils.ToCurrencyString(total, 2)))
	}

	transactionsTable.SetSelectionChangedFunc(func(row, column int) {
		if row == 0 && len(sortedTransactions) > 0 {
			transactionsTable.Select(1, column)
		}
	})
	transactionsTable.Select(1, 0).SetFixed(1, 2).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			pages.SwitchToPage("Accounts")
		}
	})

	return transactionsTable
}

func SingleAccountPage(account types.Account, portfolio types.AnalyzedPortfolio, app *tview.Application, pages *tview.Pages) tview.Primitive {
	accountTitle := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText(account.Name).
		SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorLimeGreen).Bold(true))

	valuesSection := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tview.NewTextView().SetText(fmt.Sprintf("Value: %s", utils.ToCurrencyString(portfolio.Value, 0))), 0, 1, false).
		AddItem(tview.NewTextView().SetText(fmt.Sprintf("Dividends: %s", utils.ToCurrencyString(portfolio.TotalDividends, 0))), 0, 1, false)

	yieldsSection := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tview.NewTextView().SetText(fmt.Sprintf("Dietz: %s", utils.ToYieldString(portfolio.ModifiedDietzYield))), 0, 1, false).
		AddItem(tview.NewTextView().SetText(fmt.Sprintf("Annualized: %s", utils.ToYieldString(portfolio.AnnualizedYield))), 0, 1, false)

	symbolsText := "Holdings: "
	for symbol, count := range portfolio.SymbolsCount {
		if count > 0 {
			symbolsText += fmt.Sprintf("%s(%d) ", symbol, count)
		}
	}
	symbolsSection := tview.NewTextView().SetText(symbolsText).SetWrap(true)

	head := tview.NewFlex().AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(accountTitle, 1, 1, false).
		AddItem(tview.NewTextView().SetText(fmt.Sprintf("Institution: %s", account.Institution)), 1, 1, false).
		AddItem(tview.NewTextView().SetText(fmt.Sprintf("Inception: %s", portfolio.FirstTransaction.Date.Format("2006-01-02"))), 1, 1, false).
		AddItem(valuesSection, 1, 2, false).
		AddItem(yieldsSection, 1, 2, false).
		AddItem(symbolsSection, 2, 1, false), 0, 2, false)
	head.SetBackgroundColor(tcell.ColorLightYellow)

	transactionsTable := TransactionsTable(portfolio, pages)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).SetFullScreen(true)
	layout.AddItem(head, 0, 1, false)
	layout.AddItem(transactionsTable, 0, 4, true)

	return layout
}
