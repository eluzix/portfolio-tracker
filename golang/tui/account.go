package tui

import (
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"tracker/loaders"
	"tracker/types"
	"tracker/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func fillTransactionsTable(table *tview.Table, transactions []types.Transaction, currency string, multiplier float64, theme Theme, showDividends bool) []types.Transaction {
	table.Clear()

	headerStyle := tcell.StyleDefault.
		Background(theme.HeaderBg).
		Foreground(theme.HeaderFg).Bold(true)

	table.SetCell(0, 0, tview.NewTableCell("Date").SetStyle(headerStyle).SetExpansion(1).SetAlign(tview.AlignLeft))
	table.SetCell(0, 1, tview.NewTableCell("Type").SetStyle(headerStyle).SetExpansion(1).SetAlign(tview.AlignLeft))
	table.SetCell(0, 2, tview.NewTableCell("Symbol").SetStyle(headerStyle).SetExpansion(1).SetAlign(tview.AlignLeft))
	table.SetCell(0, 3, tview.NewTableCell("Quantity").SetStyle(headerStyle).SetExpansion(1).SetAlign(tview.AlignRight))
	table.SetCell(0, 4, tview.NewTableCell("Price").SetStyle(headerStyle).SetExpansion(2).SetAlign(tview.AlignRight))
	table.SetCell(0, 5, tview.NewTableCell("Total").SetStyle(headerStyle).SetExpansion(2).SetAlign(tview.AlignRight))

	sortedTransactions := make([]types.Transaction, len(transactions))
	copy(sortedTransactions, transactions)
	sort.Slice(sortedTransactions, func(i, j int) bool {
		return sortedTransactions[i].Date.After(sortedTransactions[j].Date)
	})

	var displayedTransactions []types.Transaction
	row := 1
	for _, tx := range sortedTransactions {
		if !showDividends && (tx.Type == types.TransactionTypeDividend || tx.Type == types.TransactionTypeSplit) {
			continue
		}
		displayedTransactions = append(displayedTransactions, tx)

		table.SetCell(row, 0, tview.NewTableCell(tx.Date.Format("2006-01-02")).SetExpansion(1).SetAlign(tview.AlignLeft))
		table.SetCell(row, 1, tview.NewTableCell(string(tx.Type)).SetExpansion(1).SetAlign(tview.AlignLeft))
		table.SetCell(row, 2, tview.NewTableCell(tx.Symbol).SetExpansion(1).SetAlign(tview.AlignLeft))
		table.SetCell(row, 3, tview.NewTableCell(fmt.Sprintf("%d", tx.Quantity)).SetExpansion(1).SetAlign(tview.AlignRight))
		table.SetCell(row, 4, tview.NewTableCell(utils.ToCurrencyString(int64(tx.Pps), 2, currency, multiplier)).SetExpansion(2).SetAlign(tview.AlignRight))

		total := int64(tx.Quantity) * int64(tx.Pps)
		table.SetCell(row, 5, tview.NewTableCell(utils.ToCurrencyString(total, 2, currency, multiplier)).SetExpansion(2).SetAlign(tview.AlignRight))
		row++
	}

	table.Select(1, 0).SetFixed(1, 2)
	return displayedTransactions
}

func SingleAccountPage(db *sql.DB, analysis *types.AnalysisData, account types.Account, app *tview.Application, pages *tview.Pages) tview.Primitive {
	theme := GetTheme()

	portfolio := analysis.AccountsData[account.Id]
	currency := analysis.ExchaneSign
	rate := analysis.ExchangeRate
	multiplier := 1.0
	if currency == "₪" {
		multiplier = float64(rate)
	}

	showDividends := true
	var displayedTransactions []types.Transaction

	accountTitle := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText(account.Name).
		SetTextStyle(tcell.StyleDefault.Foreground(theme.Positive).Bold(true))

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

	tagsText := "Tags: "
	if len(account.Tags) > 0 {
		tagsText += strings.Join(account.Tags, ", ")
	} else {
		tagsText += "-"
	}
	tagsSection := tview.NewTextView().SetText(tagsText).SetWrap(true)

	head := tview.NewFlex().AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(accountTitle, 1, 1, false).
		AddItem(tview.NewTextView().SetText(fmt.Sprintf("Institution: %s", account.Institution)), 1, 1, false).
		AddItem(tview.NewTextView().SetText(fmt.Sprintf("Inception: %s", portfolio.FirstTransaction.Date.Format("2006-01-02"))), 1, 1, false).
		AddItem(valuesSection, 1, 2, false).
		AddItem(yieldsSection, 1, 2, false).
		AddItem(symbolsSection, 2, 1, false).
		AddItem(tagsSection, 1, 1, false), 0, 2, false)
	head.SetBackgroundColor(theme.HeaderBg)

	transactionsTable := tview.NewTable().SetContent(nil)
	transactionsTable.SetBorder(true).SetBorderColor(theme.Border)
	transactionsTable.SetSelectable(true, false)
	transactionsTable.SetSeparator('|').SetBorderPadding(2, 2, 3, 3)
	transactionsTable.SetSelectedStyle(tcell.StyleDefault.Background(theme.SelectedBg).Foreground(theme.SelectedFg))

	displayedTransactions = fillTransactionsTable(transactionsTable, portfolio.Transactions, currency, multiplier, theme, showDividends)

	addButton := tview.NewButton("Add Transaction")
	addButton.SetStyle(tcell.StyleDefault.Background(theme.ButtonBg).Foreground(theme.ButtonFg))
	addButton.SetSelectedFunc(func() {
		showAddTransactionModal(db, app, pages, account, theme)
	})

	toggleButton := tview.NewButton("Hide Dividends")
	toggleButton.SetStyle(tcell.StyleDefault.Background(theme.ButtonBg).Foreground(theme.ButtonFg))
	toggleButton.SetSelectedFunc(func() {
		showDividends = !showDividends
		if showDividends {
			toggleButton.SetLabel("Hide Dividends")
		} else {
			toggleButton.SetLabel("Show Dividends")
		}
		displayedTransactions = fillTransactionsTable(transactionsTable, portfolio.Transactions, currency, multiplier, theme, showDividends)
		app.Sync()
	})

	buttonSection := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(addButton, 20, 1, false).
		AddItem(tview.NewTextView().SetText(" "), 2, 0, false).
		AddItem(toggleButton, 20, 1, false).
		AddItem(nil, 0, 1, false)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).SetFullScreen(true)
	layout.AddItem(head, 0, 1, false)
	layout.AddItem(buttonSection, 1, 1, false)
	layout.AddItem(transactionsTable, 0, 4, true)

	transactionsTable.SetSelectionChangedFunc(func(row, column int) {
		if row == 0 && len(displayedTransactions) > 0 {
			transactionsTable.Select(1, column)
		}
	})

	focusItem := 0
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			if focusItem == 0 {
				app.SetFocus(addButton)
				focusItem = 1
			} else if focusItem == 1 {
				app.SetFocus(toggleButton)
				focusItem = 2
			} else {
				app.SetFocus(transactionsTable)
				focusItem = 0
			}
			return nil
		}
		if event.Key() == tcell.KeyCtrlD {
			row, _ := transactionsTable.GetSelection()
			if row >= 1 && row <= len(displayedTransactions) {
				transaction := displayedTransactions[row-1]
				if transaction.Type == types.TransactionTypeBuy || transaction.Type == types.TransactionTypeSell {
					showDeleteConfirmation(db, app, pages, transaction, currency, multiplier, theme)
				}
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

	return layout
}

func showAddTransactionModal(db *sql.DB, app *tview.Application, pages *tview.Pages, account types.Account, theme Theme) {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle("Add Transaction").SetTitleAlign(tview.AlignLeft)
	form.SetBackgroundColor(theme.ModalBg)
	form.SetButtonBackgroundColor(theme.ButtonBg)
	form.SetButtonTextColor(theme.ButtonFg)

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
			showAbandonConfirmation(app, pages, theme)
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

func showAbandonConfirmation(app *tview.Application, pages *tview.Pages, theme Theme) {
	confirmForm := tview.NewForm()
	confirmForm.SetBorder(true).SetTitle("Abandon Changes?").SetTitleAlign(tview.AlignLeft)
	confirmForm.SetBackgroundColor(theme.ModalBg)
	confirmForm.SetButtonBackgroundColor(theme.ButtonBg)
	confirmForm.SetButtonTextColor(theme.ButtonFg)

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

func showDeleteConfirmation(db *sql.DB, app *tview.Application, pages *tview.Pages, transaction types.Transaction, currency string, multiplier float64, theme Theme) {
	deleteForm := tview.NewForm()
	deleteForm.SetBorder(true).SetTitle("Delete Transaction?").SetTitleAlign(tview.AlignLeft)
	deleteForm.SetBackgroundColor(theme.ModalBg)
	deleteForm.SetButtonBackgroundColor(theme.ButtonBg)
	deleteForm.SetButtonTextColor(theme.ButtonFg)

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
