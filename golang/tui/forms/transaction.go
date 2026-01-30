package forms

import (
	"fmt"
	"strconv"
	"time"

	"tracker/types"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type TransactionForm struct {
	form      *huh.Form
	accountId string
	width     int
	height    int
	completed bool
	cancelled bool
	result    types.Transaction
	styles    TransactionFormStyles
}

type TransactionFormStyles struct {
	Container lipgloss.Style
	Title     lipgloss.Style
}

func DefaultTransactionFormStyles() TransactionFormStyles {
	return TransactionFormStyles{
		Container: lipgloss.NewStyle().
			Background(lipgloss.Color("#24283b")).
			Foreground(lipgloss.Color("#e0e6f0")).
			Padding(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#bb9af7")),
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff9e64")).
			Bold(true).
			MarginBottom(1),
	}
}

type TransactionFormResult struct {
	Transaction types.Transaction
	Cancelled   bool
}

func NewTransactionForm(accountId string) TransactionForm {
	var (
		dateStr     = time.Now().Format("2006-01-02")
		txType      = "Buy"
		symbol      = ""
		quantityStr = ""
		priceStr    = ""
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("date").
				Title("Date (YYYY-MM-DD)").
				Value(&dateStr).
				Validate(func(s string) error {
					_, err := time.Parse("2006-01-02", s)
					if err != nil {
						return fmt.Errorf("invalid date format")
					}
					return nil
				}),

			huh.NewSelect[string]().
				Key("type").
				Title("Transaction Type").
				Options(
					huh.NewOption("Buy", "Buy"),
					huh.NewOption("Sell", "Sell"),
				).
				Value(&txType),

			huh.NewInput().
				Key("symbol").
				Title("Symbol").
				Value(&symbol).
				Validate(func(s string) error {
					if len(s) == 0 {
						return fmt.Errorf("symbol is required")
					}
					return nil
				}),

			huh.NewInput().
				Key("quantity").
				Title("Quantity").
				Value(&quantityStr).
				Validate(func(s string) error {
					if _, err := strconv.ParseInt(s, 10, 32); err != nil {
						return fmt.Errorf("must be a number")
					}
					return nil
				}),

			huh.NewInput().
				Key("price").
				Title("Price per Share").
				Value(&priceStr).
				Validate(func(s string) error {
					if _, err := strconv.ParseFloat(s, 64); err != nil {
						return fmt.Errorf("must be a number")
					}
					return nil
				}),

			huh.NewConfirm().
				Key("confirm").
				Title("Save transaction?").
				Affirmative("Save").
				Negative("Cancel"),
		),
	).WithTheme(getFormTheme()).WithShowHelp(true).WithShowErrors(true).WithKeyMap(getFormKeyMap())

	return TransactionForm{
		form:      form,
		accountId: accountId,
		styles:    DefaultTransactionFormStyles(),
	}
}

func (f *TransactionForm) SetSize(width, height int) {
	f.width = width
	f.height = height
	f.form.WithWidth(width - 10)
	f.form.WithHeight(height - 10)
}

func (f *TransactionForm) Completed() bool {
	return f.completed
}

func (f *TransactionForm) Cancelled() bool {
	return f.cancelled
}

func (f *TransactionForm) Result() types.Transaction {
	return f.result
}

func (f *TransactionForm) Init() tea.Cmd {
	return f.form.Init()
}

func (f *TransactionForm) Update(msg tea.Msg) (TransactionForm, tea.Cmd) {
	form, cmd := f.form.Update(msg)
	if ff, ok := form.(*huh.Form); ok {
		f.form = ff
	}

	if f.form.State == huh.StateCompleted {
		f.completed = true
		confirm := f.form.GetBool("confirm")
		if !confirm {
			f.cancelled = true
		} else {
			f.result = f.buildTransaction()
		}
	}

	return *f, cmd
}

func (f *TransactionForm) buildTransaction() types.Transaction {
	dateStr := f.form.GetString("date")
	txType := f.form.GetString("type")
	symbol := f.form.GetString("symbol")
	quantityStr := f.form.GetString("quantity")
	priceStr := f.form.GetString("price")

	date, _ := time.Parse("2006-01-02", dateStr)
	quantity, _ := strconv.ParseInt(quantityStr, 10, 32)
	price, _ := strconv.ParseFloat(priceStr, 64)
	priceInCents := int32(price * 100)

	var transactionType types.TransactionType
	if txType == "Buy" {
		transactionType = types.TransactionTypeBuy
	} else {
		transactionType = types.TransactionTypeSell
	}

	return types.Transaction{
		AccountId: f.accountId,
		Symbol:    symbol,
		Date:      date,
		Type:      transactionType,
		Quantity:  int32(quantity),
		Pps:       priceInCents,
	}
}

func (f TransactionForm) View() string {
	title := f.styles.Title.Render("Add Transaction")
	formView := f.form.View()

	content := lipgloss.JoinVertical(lipgloss.Left, title, formView)
	modal := f.styles.Container.Render(content)

	return lipgloss.Place(f.width, f.height,
		lipgloss.Center, lipgloss.Center,
		modal,
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#1a1b26")),
	)
}

func getFormTheme() *huh.Theme {
	t := huh.ThemeCatppuccin()

	t.Focused.Base = t.Focused.Base.
		BorderForeground(lipgloss.Color("#bb9af7"))
	t.Focused.Title = t.Focused.Title.
		Foreground(lipgloss.Color("#7dcfff"))
	t.Focused.Description = t.Focused.Description.
		Foreground(lipgloss.Color("#737aa2"))
	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.
		Foreground(lipgloss.Color("#ff9e64"))
	t.Focused.SelectSelector = t.Focused.SelectSelector.
		Foreground(lipgloss.Color("#ff9e64"))

	t.Blurred.Base = t.Blurred.Base.
		BorderForeground(lipgloss.Color("#414868"))

	return t
}

func getFormKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	km.Input.Next = key.NewBinding(key.WithKeys("tab", "enter"))
	km.Input.Prev = key.NewBinding(key.WithKeys("shift+tab"))
	km.Select.Next = key.NewBinding(key.WithKeys("tab", "enter"))
	km.Select.Prev = key.NewBinding(key.WithKeys("shift+tab"))
	km.Text.Next = key.NewBinding(key.WithKeys("tab"))
	km.Text.Prev = key.NewBinding(key.WithKeys("shift+tab"))
	km.Confirm.Next = key.NewBinding(key.WithKeys("tab", "enter"))
	km.Confirm.Prev = key.NewBinding(key.WithKeys("shift+tab"))
	return km
}
