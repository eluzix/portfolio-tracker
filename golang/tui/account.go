package tui

import (
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"time"
	"tracker/loaders"
	"tracker/types"
	"tracker/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TransactionsTable(db *sql.DB, analysis *types.AnalysisData, accountId string, app *tview.Application, pages *tview.Pages) *tview.Table {
	portfolio := analysis.AccountsData[accountId]
	currency := analysis.ExchaneSign
	rate := analysis.ExchangeRate
	multiplier := 1.0
	if currency == "₪" {
		multiplier = float64(rate)
	}
	transactionsTable := tview.NewTable().SetContent(nil)
	transactionsTable.SetBorder(true).SetBorderColor(tcell.ColorGreenYellow)
	transactionsTable.SetSelectable(true, false)
	transactionsTable.SetSeparator('|').SetBorderPadding(2, 2, 3, 3)

	headerStyle := tcell.StyleDefault.
		Background(tcell.ColorOrangeRed).
		Foreground(tcell.ColorGreenYellow).Bold(true)

	transactionsTable.SetCell(0, 0, tview.NewTableCell("Date").SetStyle(headerStyle).SetExpansion(1).SetAlign(tview.AlignLeft))
	transactionsTable.SetCell(0, 1, tview.NewTableCell("Type").SetStyle(headerStyle).SetExpansion(1).SetAlign(tview.AlignLeft))
	transactionsTable.SetCell(0, 2, tview.NewTableCell("Symbol").SetStyle(headerStyle).SetExpansion(1).SetAlign(tview.AlignLeft))
	transactionsTable.SetCell(0, 3, tview.NewTableCell("Quantity").SetStyle(headerStyle).SetExpansion(1).SetAlign(tview.AlignRight))
	transactionsTable.SetCell(0, 4, tview.NewTableCell("Price").SetStyle(headerStyle).SetExpansion(2).SetAlign(tview.AlignRight))
	transactionsTable.SetCell(0, 5, tview.NewTableCell("Total").SetStyle(headerStyle).SetExpansion(2).SetAlign(tview.AlignRight))

	sortedTransactions := make([]types.Transaction, len(portfolio.Transactions))
	copy(sortedTransactions, portfolio.Transactions)
	sort.Slice(sortedTransactions, func(i, j int) bool {
		return sortedTransactions[i].Date.After(sortedTransactions[j].Date)
	})

	for i, tx := range sortedTransactions {
		row := i + 1
		transactionsTable.SetCell(row, 0, tview.NewTableCell(tx.Date.Format("2006-01-02")).SetExpansion(1).SetAlign(tview.AlignLeft))
		transactionsTable.SetCell(row, 1, tview.NewTableCell(string(tx.Type)).SetExpansion(1).SetAlign(tview.AlignLeft))
		transactionsTable.SetCell(row, 2, tview.NewTableCell(tx.Symbol).SetExpansion(1).SetAlign(tview.AlignLeft))
		transactionsTable.SetCell(row, 3, tview.NewTableCell(fmt.Sprintf("%d", tx.Quantity)).SetExpansion(1).SetAlign(tview.AlignRight))
		transactionsTable.SetCell(row, 4, tview.NewTableCell(utils.ToCurrencyString(int64(tx.Pps), 2, currency, multiplier)).SetExpansion(2).SetAlign(tview.AlignRight))

		total := int64(tx.Quantity) * int64(tx.Pps)
		transactionsTable.SetCell(row, 5, tview.NewTableCell(utils.ToCurrencyString(total, 2, currency, multiplier)).SetExpansion(2).SetAlign(tview.AlignRight))
	}

	transactionsTable.SetSelectionChangedFunc(func(row, column int) {
		if row == 0 && len(sortedTransactions) > 0 {
			transactionsTable.Select(1, column)
		}
	})

	transactionsTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlD {
		row, _ := transactionsTable.GetSelection()
		if row >= 1 && row <= len(sortedTransactions) {
		transaction := sortedTransactions[row-1]
		showDeleteConfirmation(db, app, pages, transaction, currency, multiplier)
		}
		return nil
		}
		if event.Key() == tcell.KeyEscape {
			pages.SwitchToPage("Accounts")
			return nil
		}
		return event
	})

	app.SetFocus(transactionsTable)

	transactionsTable.Select(1, 0).SetFixed(1, 2)

	return transactionsTable
}

func SingleAccountPage(db *sql.DB, analysis *types.AnalysisData, account types.Account, app *tview.Application, pages *tview.Pages) tview.Primitive {
	portfolio := analysis.AccountsData[account.Id]
	currency := analysis.ExchaneSign
	rate := analysis.ExchangeRate
	multiplier := 1.0
	if currency == "₪" {
		multiplier = float64(rate)
	}
	accountTitle := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText(account.Name).
		SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorLimeGreen).Bold(true))

	valuesSection := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tview.NewTextView().SetText(fmt.Sprintf("Value: %s", utils.ToCurrencyString(portfolio.Value, 0, currency, multiplier))), 0, 1, false).
		AddItem(tview.NewTextView().SetText(fmt.Sprintf("Dividends: %s", utils.ToCurrencyString(portfolio.TotalDividends, 0, currency, multiplier))), 0, 1, false)

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

	transactionsTable := TransactionsTable(db, analysis, account.Id, app, pages)

	addButton := tview.NewButton("Add Transaction")
	addButton.SetBackgroundColor(tcell.ColorDarkGreen)
	addButton.SetSelectedFunc(func() {
		showAddTransactionModal(db, app, pages, account)
	})

	buttonSection := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(addButton, 20, 1, false).
		AddItem(nil, 0, 1, false)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).SetFullScreen(true)
	layout.AddItem(head, 0, 1, false)
	layout.AddItem(buttonSection, 1, 1, false)
	layout.AddItem(transactionsTable, 0, 4, true)

	// Set up tab navigation between button and table
	addButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			app.SetFocus(transactionsTable)
			return nil
		}
		if event.Key() == tcell.KeyEscape {
			pages.SwitchToPage("Accounts")
			return nil
		}
		return event
	})

	transactionsTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlD {
			row, _ := transactionsTable.GetSelection()
			if row >= 1 && row <= len(portfolio.Transactions) {
				sortedTransactions := make([]types.Transaction, len(portfolio.Transactions))
				copy(sortedTransactions, portfolio.Transactions)
				sort.Slice(sortedTransactions, func(i, j int) bool {
					return sortedTransactions[i].Date.After(sortedTransactions[j].Date)
				})
				transaction := sortedTransactions[row-1]
				if transaction.Type == types.TransactionTypeBuy || transaction.Type == types.TransactionTypeSell {
					showDeleteConfirmation(db, app, pages, transaction, currency, multiplier)
				}
			}
			return nil
		}
		if event.Key() == tcell.KeyTab {
			app.SetFocus(addButton)
			return nil
		}
		if event.Key() == tcell.KeyEscape {
			pages.SwitchToPage("Accounts")
			return nil
		}
		return event
	})

	return layout
}

func showAddTransactionModal(db *sql.DB, app *tview.Application, pages *tview.Pages, account types.Account) {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle("Add Transaction").SetTitleAlign(tview.AlignLeft)
	form.SetBackgroundColor(tcell.ColorBlack)

	form.AddInputField("Date (YYYY-MM-DD)", time.Now().Format("2006-01-02"), 20, nil, nil)
	form.AddDropDown("Type", []string{"Buy", "Sell"}, 0, nil)
	form.AddInputField("Symbol", "", 20, nil, nil)
	form.AddInputField("Quantity", "", 20, nil, nil)
	form.AddInputField("Price per Share", "", 20, nil, nil)

	form.AddButton("Save", func() {
		dateStr := form.GetFormItem(0).(*tview.InputField).GetText()
		_, typeStr := form.GetFormItem(1).(*tview.DropDown).GetCurrentOption()
		symbol := form.GetFormItem(2).(*tview.InputField).GetText()
		quantityStr := form.GetFormItem(3).(*tview.InputField).GetText()
		priceStr := form.GetFormItem(4).(*tview.InputField).GetText()

		date := utils.StringToDate(dateStr)

		var transactionType types.TransactionType
		if typeStr == "Buy" {
			transactionType = types.TransactionTypeBuy
		} else {
			transactionType = types.TransactionTypeSell
		}

		quantity, _ := strconv.ParseInt(quantityStr, 10, 32)
		price, _ := strconv.ParseFloat(priceStr, 64)
		priceInCents := int32(price * 100)

		transaction := types.Transaction{
			Id:        "",
			AccountId: account.Id,
			Symbol:    symbol,
			Date:      date,
			Type:      transactionType,
			Quantity:  int32(quantity),
			Pps:       priceInCents,
		}

		err := loaders.AddTransaction(db, transaction)
		if err != nil {
			panic(err)
		}

		pages.RemovePage("addTransaction")
	})
	form.AddButton("Cancel", func() {
		pages.RemovePage("addTransaction")
	})

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			showAbandonConfirmation(app, pages)
			return nil
		}
		return event
	})

	// Create modal with centered form
	modal := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(form, 50, 1, true).
			AddItem(nil, 0, 1, false), 15, 1, true).
		AddItem(nil, 0, 1, false)

	pages.AddPage("addTransaction", modal, true, true)
	app.SetFocus(form)
}

func showAbandonConfirmation(app *tview.Application, pages *tview.Pages) {
	confirmForm := tview.NewForm()
	confirmForm.SetBorder(true).SetTitle("Abandon Changes?").SetTitleAlign(tview.AlignLeft)
	confirmForm.SetBackgroundColor(tcell.ColorBlack)

	confirmForm.AddTextView("", "Are you sure you want to abandon your changes?", 40, 3, true, false)

	confirmForm.AddButton("Yes, abandon", func() {
		pages.RemovePage("abandonConfirm")
		pages.RemovePage("addTransaction")
	})
	confirmForm.AddButton("No, continue editing", func() {
		pages.RemovePage("abandonConfirm")
	})

	confirmForm.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			pages.RemovePage("abandonConfirm")
			return nil
		}
		return event
	})

	confirmModal := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(confirmForm, 50, 1, true).
			AddItem(nil, 0, 1, false), 10, 1, true).
		AddItem(nil, 0, 1, false)

	pages.AddPage("abandonConfirm", confirmModal, true, true)
	app.SetFocus(confirmForm)
}

func showDeleteConfirmation(db *sql.DB, app *tview.Application, pages *tview.Pages, transaction types.Transaction, currency string, multiplier float64) {
	deleteForm := tview.NewForm()
	deleteForm.SetBorder(true).SetTitle("Delete Transaction?").SetTitleAlign(tview.AlignLeft)
	deleteForm.SetBackgroundColor(tcell.ColorBlack)

	message := fmt.Sprintf("Are you sure you want to delete this transaction?\n\n%s %s - %d shares @ %s",
		transaction.Date.Format("2006-01-02"),
		transaction.Symbol,
		transaction.Quantity,
		utils.ToCurrencyString(int64(transaction.Pps), 2, currency, multiplier))

	deleteForm.AddTextView("", message, 50, 5, true, false)

	deleteForm.AddButton("Yes, delete", func() {
		err := loaders.DeleteTransaction(db, transaction.Id)
		if err != nil {
			panic(fmt.Sprintf("Error deleting transaction: %s", err.Error()))
		}
		pages.RemovePage("deleteConfirm")
		pages.SwitchToPage("Accounts")
	})
	deleteForm.AddButton("No, cancel", func() {
		pages.RemovePage("deleteConfirm")
	})

	deleteForm.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			pages.RemovePage("deleteConfirm")
			return nil
		}
		return event
	})

	deleteModal := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(deleteForm, 60, 1, true).
			AddItem(nil, 0, 1, false), 12, 1, true).
		AddItem(nil, 0, 1, false)

	pages.AddPage("deleteConfirm", deleteModal, true, true)
	app.SetFocus(deleteForm)
}
