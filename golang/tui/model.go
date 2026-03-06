package tui

import (
	"database/sql"
	"fmt"

	"tracker/config"
	"tracker/llm"
	"tracker/loaders"
	"tracker/market"
	"tracker/portfolio"
	"tracker/tui/components"
	"tracker/tui/forms"
	"tracker/tui/views"
	"tracker/types"
	"tracker/utils"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type View int

const (
	ViewLoading View = iota
	ViewAccounts
	ViewAccountDetail
)

type Model struct {
	db                *sql.DB
	width             int
	height            int
	view              View
	loading           bool
	showHelp          bool
	modalType         ModalType
	spinner           spinner.Model
	help              help.Model
	header            components.Header
	statusBar         components.StatusBar
	accountsView      views.AccountsView
	accountDetailView views.AccountDetailView
	insightsView      *views.InsightsView
	transactionForm   forms.TransactionForm
	confirmDialog     forms.ConfirmDialog
	pendingDeleteTx   *types.Transaction
	styles            Styles
	accounts          *[]types.Account
	accountsData      map[string]types.AnalyzedPortfolio
	allPortfolio      types.AnalyzedPortfolio
	selectedAccount   types.Account
	currencySymbol    string
	exchangeRate      float64
	tagFilter         string
	tags              []string
	tagIndex          int
	showDividends     bool
	dividendTaxRate   float64
	statusText        string
	err               error
}

func NewModel(db *sql.DB, cfg config.AppConfig) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff9e64"))

	h := help.New()
	h.Styles.ShortKey = AppStyles.HelpKey
	h.Styles.ShortDesc = AppStyles.HelpDesc
	h.Styles.FullKey = AppStyles.HelpKey
	h.Styles.FullDesc = AppStyles.HelpDesc

	header := components.NewHeader()
	header.SetTitle("Portfolio Tracker")

	statusBar := components.NewStatusBar()
	statusBar.SetMode("LOADING")
	statusBar.SetHint("? help  q quit")
	statusBar.SetLoading(true)

	return Model{
		db:              db,
		view:            ViewLoading,
		loading:         true,
		spinner:         s,
		help:            h,
		header:          header,
		statusBar:       statusBar,
		styles:          AppStyles,
		currencySymbol:  market.CurrencySymbolUSD,
		exchangeRate:    1.0,
		tagFilter:       "All",
		tags:            []string{"All"},
		showDividends:   true,
		dividendTaxRate: cfg.DividendTaxRate,
		modalType:       ModalNone,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.loadData(),
	)
}

func (m Model) loadData() tea.Cmd {
	return func() tea.Msg {
		accounts, err := loaders.UserAccounts(m.db)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		accountsData := make(map[string]types.AnalyzedPortfolio, len(*accounts))
		for _, ac := range *accounts {
			data, err := portfolio.LoadAndAnalyze(m.db, ac)
			if err != nil {
				return ErrorMsg{Err: err}
			}
			accountsData[ac.Id] = data
		}

		var accountIds []string
		for _, ac := range *accounts {
			accountIds = append(accountIds, ac.Id)
		}
		allPortfolio, _ := portfolio.LoadAndAnalyzeAccounts(m.db, accountIds)

		return DataLoadedMsg{
			Accounts:     accounts,
			AccountsData: accountsData,
			AllPortfolio: allPortfolio,
		}
	}
}

func (m *Model) loadExchangeRate(currency string) tea.Cmd {
	m.statusBar.SetLoading(true)
	m.statusBar.SetStatus("Loading exchange rate...")
	return func() tea.Msg {
		if currency == "USD" {
			return CurrencyChangedMsg{Symbol: market.CurrencySymbolUSD, ExchangeRate: 1.0}
		}
		rate := loaders.CurrencyExchangeRate(m.db, "ILS")
		return CurrencyChangedMsg{Symbol: market.CurrencySymbolILS, ExchangeRate: rate}
	}
}

func (m *Model) reloadAccountData() tea.Cmd {
	return func() tea.Msg {
		accounts, err := loaders.UserAccounts(m.db)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		accountsData := make(map[string]types.AnalyzedPortfolio, len(*accounts))
		for _, ac := range *accounts {
			data, err := portfolio.LoadAndAnalyze(m.db, ac)
			if err != nil {
				return ErrorMsg{Err: err}
			}
			accountsData[ac.Id] = data
		}

		var accountIds []string
		for _, ac := range *accounts {
			accountIds = append(accountIds, ac.Id)
		}
		allPortfolio, _ := portfolio.LoadAndAnalyzeAccounts(m.db, accountIds)

		return DataLoadedMsg{
			Accounts:     accounts,
			AccountsData: accountsData,
			AllPortfolio: allPortfolio,
		}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if m.modalType != ModalNone {
		return m.updateModal(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		m.header.SetWidth(msg.Width)
		m.statusBar.SetWidth(msg.Width)
		if m.accounts != nil {
			m.accountsView.SetSize(msg.Width, msg.Height-4)
			m.accountDetailView.SetSize(msg.Width, msg.Height-4)
		}

	case tea.KeyMsg:

		if m.showHelp {
			if key.Matches(msg, Keys.Help) || key.Matches(msg, Keys.Back) {
				m.showHelp = false
			}
			return m, nil
		}

		if key.Matches(msg, Keys.Quit) {
			return m, tea.Quit
		}
		if key.Matches(msg, Keys.Help) {
			m.showHelp = true
			return m, nil
		}

		switch m.view {
		case ViewAccounts:
			return m.updateAccountsView(msg)
		case ViewAccountDetail:
			return m.updateAccountDetailView(msg)
		}

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case DataLoadedMsg:
		m.loading = false
		m.accounts = msg.Accounts
		m.accountsData = msg.AccountsData
		m.allPortfolio = msg.AllPortfolio
		m.tags = collectUniqueTags(msg.Accounts)

		if m.view == ViewLoading {
			m.view = ViewAccounts
		}

		m.accountsView = views.NewAccountsView(m.accounts, m.accountsData, m.allPortfolio)
		m.accountsView.SetSize(m.width, m.height-4)
		m.accountsView.SetCurrency(m.currencySymbol, m.exchangeRate)

		if m.selectedAccount.Id != "" {
			m.accountDetailView = views.NewAccountDetailView(m.selectedAccount, m.accountsData[m.selectedAccount.Id], m.dividendTaxRate)
			m.accountDetailView.SetSize(m.width, m.height-4)
			m.accountDetailView.SetCurrency(m.currencySymbol, m.exchangeRate)
		}

		m.statusText = fmt.Sprintf("%d accounts loaded", len(*msg.Accounts))
		m.statusBar.SetLoading(false)
		m.statusBar.SetStatus(m.statusText)
		if m.view == ViewAccounts {
			m.statusBar.SetMode("ACCOUNTS")
			m.header.SetSubtitle("All Accounts")
		}

	case CurrencyChangedMsg:
		m.currencySymbol = msg.Symbol
		m.exchangeRate = msg.ExchangeRate
		m.accountsView.SetCurrency(msg.Symbol, msg.ExchangeRate)
		m.accountDetailView.SetCurrency(msg.Symbol, msg.ExchangeRate)
		m.statusBar.SetLoading(false)
		m.statusBar.SetStatus("Currency: " + msg.Symbol)

	case InsightsLoadedMsg:
		iv := views.NewInsightsView(msg.Title, msg.Content)
		iv.SetSize(m.width, m.height)
		m.insightsView = &iv
		m.modalType = ModalInsights
		m.statusBar.SetLoading(false)
		m.statusBar.SetStatus("")

	case InsightsErrorMsg:
		iv := views.NewInsightsView(msg.Title, "Error generating insights:\n\n"+msg.Error)
		iv.SetSize(m.width, m.height)
		m.insightsView = &iv
		m.modalType = ModalInsights
		m.statusBar.SetLoading(false)
		m.statusBar.SetStatus("")

	case TransactionAddedMsg:
		m.statusBar.SetStatus("Transaction added")
		return m, m.reloadAccountData()

	case TransactionDeletedMsg:
		m.statusBar.SetStatus("Transaction deleted")
		return m, m.reloadAccountData()

	case ErrorMsg:
		m.err = msg.Err
		m.loading = false
		m.statusText = "Error: " + msg.Error()
		m.statusBar.SetStatus("Error: " + msg.Error())
		m.statusBar.SetLoading(false)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) updateModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.modalType == ModalAddTransaction {
			m.transactionForm.SetSize(msg.Width, msg.Height)
		}
		if m.modalType == ModalDeleteConfirm || m.modalType == ModalAbandonConfirm {
			m.confirmDialog.SetSize(msg.Width, msg.Height)
		}
		if m.modalType == ModalInsights && m.insightsView != nil {
			m.insightsView.SetSize(msg.Width, msg.Height)
		}
		return m, nil

	case tea.KeyMsg:
		if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
			if m.modalType == ModalAddTransaction {
				m.modalType = ModalNone
				m.statusBar.SetStatus("Cancelled")
				return m, nil
			}
			if m.modalType == ModalInsights {
				m.modalType = ModalNone
				m.insightsView = nil
				return m, nil
			}
		}

		if m.modalType == ModalInsights {
			switch {
			case key.Matches(msg, Keys.Back):
				m.modalType = ModalNone
				m.insightsView = nil
				return m, nil
			case key.Matches(msg, Keys.Up):
				if m.insightsView != nil {
					m.insightsView.ScrollUp()
				}
				return m, nil
			case key.Matches(msg, Keys.Down):
				if m.insightsView != nil {
					m.insightsView.ScrollDown()
				}
				return m, nil
			}
		}
	}

	switch m.modalType {
	case ModalAddTransaction:
		m.transactionForm, cmd = m.transactionForm.Update(msg)
		if m.transactionForm.Completed() {
			if m.transactionForm.Cancelled() {
				m.modalType = ModalNone
				m.statusBar.SetStatus("Cancelled")
			} else {
				tx := m.transactionForm.Result()
				err := loaders.AddTransaction(m.db, tx)
				if err != nil {
					m.statusBar.SetStatus("Error: " + err.Error())
				} else {
					m.modalType = ModalNone
					return m, func() tea.Msg { return TransactionAddedMsg{Transaction: tx} }
				}
			}
		}
		return m, cmd

	case ModalDeleteConfirm:
		m.confirmDialog, cmd = m.confirmDialog.Update(msg)
		if m.confirmDialog.Completed() {
			if m.confirmDialog.Confirmed() && m.pendingDeleteTx != nil {
				err := loaders.DeleteTransaction(m.db, m.pendingDeleteTx.Id)
				if err != nil {
					m.statusBar.SetStatus("Error: " + err.Error())
				} else {
					m.modalType = ModalNone
					m.pendingDeleteTx = nil
					return m, func() tea.Msg { return TransactionDeletedMsg{TransactionID: m.pendingDeleteTx.Id} }
				}
			} else {
				m.modalType = ModalNone
				m.pendingDeleteTx = nil
				m.statusBar.SetStatus("Cancelled")
			}
		}
		return m, cmd

	case ModalAbandonConfirm:
		m.confirmDialog, cmd = m.confirmDialog.Update(msg)
		if m.confirmDialog.Completed() {
			m.modalType = ModalNone
			if m.confirmDialog.Confirmed() {
				m.statusBar.SetStatus("Changes abandoned")
			}
		}
		return m, cmd

	case ModalInsights:
		return m, nil
	}

	return m, nil
}

func (m Model) updateAccountsView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch {
	case key.Matches(msg, Keys.Enter):
		if account := m.accountsView.SelectedAccount(); account != nil {
			m.selectedAccount = *account
			m.view = ViewAccountDetail
			m.accountDetailView = views.NewAccountDetailView(*account, m.accountsData[account.Id], m.dividendTaxRate)
			m.accountDetailView.SetSize(m.width, m.height-4)
			m.accountDetailView.SetCurrency(m.currencySymbol, m.exchangeRate)
			m.header.SetSubtitle(account.Name)
			m.statusBar.SetMode("DETAIL")
			m.statusBar.SetStatus("")
		}
		return m, nil

	case key.Matches(msg, Keys.CurrencyUSD):
		cmd = m.loadExchangeRate("USD")
		return m, cmd

	case key.Matches(msg, Keys.CurrencyNIS):
		cmd = m.loadExchangeRate("NIS")
		return m, cmd

	case key.Matches(msg, Keys.CycleTag):
		newTag := m.accountsView.CycleTag()
		m.tagFilter = newTag
		m.statusBar.SetStatus("Tag: " + newTag)

		filteredIds := m.getFilteredAccountIds()
		allPortfolio, _ := portfolio.LoadAndAnalyzeAccounts(m.db, filteredIds)
		m.allPortfolio = allPortfolio
		m.accountsView.SetAllPortfolio(allPortfolio)
		return m, nil

	case key.Matches(msg, Keys.Summarize):
		m.statusBar.SetLoading(true)
		m.statusBar.SetStatus("Generating portfolio insights...")
		return m, m.getPortfolioInsights()

	default:
		m.accountsView, cmd = m.accountsView.Update(msg)
		return m, cmd
	}
}

func (m Model) updateAccountDetailView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch {
	case key.Matches(msg, Keys.Back):
		m.view = ViewAccounts
		m.selectedAccount = types.Account{}
		m.header.SetSubtitle("All Accounts")
		m.statusBar.SetMode("ACCOUNTS")
		m.statusBar.SetStatus("")
		m.accountsView.Focus()
		return m, nil

	case key.Matches(msg, Keys.NewTx):
		m.modalType = ModalAddTransaction
		m.transactionForm = forms.NewTransactionForm(m.selectedAccount.Id)
		m.transactionForm.SetSize(m.width, m.height)
		return m, m.transactionForm.Init()

	case key.Matches(msg, Keys.DeleteTx):
		if tx := m.accountDetailView.SelectedTransaction(); tx != nil {
			if tx.Type == types.TransactionTypeBuy || tx.Type == types.TransactionTypeSell {
				m.pendingDeleteTx = tx
				m.modalType = ModalDeleteConfirm
				message := fmt.Sprintf("%s %s - %d shares @ %s",
					tx.Date.Format("2006-01-02"),
					tx.Symbol,
					tx.Quantity,
					utils.ToCurrencyString(int64(tx.Pps), 2, m.currencySymbol, m.exchangeRate),
				)
				m.confirmDialog = forms.NewDeleteConfirmDialog(message)
				m.confirmDialog.SetSize(m.width, m.height)
			}
		}
		return m, nil

	case key.Matches(msg, Keys.ToggleDivs):
		showing := m.accountDetailView.ToggleDividends()
		if showing {
			m.statusBar.SetStatus("Showing all transactions")
		} else {
			m.statusBar.SetStatus("Hiding dividends/splits")
		}
		return m, nil

	case key.Matches(msg, Keys.CurrencyUSD):
		cmd = m.loadExchangeRate("USD")
		return m, cmd

	case key.Matches(msg, Keys.CurrencyNIS):
		cmd = m.loadExchangeRate("NIS")
		return m, cmd

	case key.Matches(msg, Keys.Summarize):
		m.statusBar.SetLoading(true)
		m.statusBar.SetStatus("Generating account insights...")
		return m, m.getAccountInsights()

	default:
		m.accountDetailView, cmd = m.accountDetailView.Update(msg)
		return m, cmd
	}
}

func (m Model) getFilteredAccountIds() []string {
	var ids []string
	for _, ac := range *m.accounts {
		if m.tagFilter == "All" || hasTag(ac.Tags, m.tagFilter) {
			ids = append(ids, ac.Id)
		}
	}
	return ids
}

func hasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	if m.modalType != ModalNone {
		return m.viewModal()
	}

	if m.showHelp {
		return views.RenderHelpOverlay(m.width, m.height, Keys.FullHelp())
	}

	var content string

	switch m.view {
	case ViewLoading:
		content = m.viewLoading()
	case ViewAccounts:
		content = m.viewAccountsPage()
	case ViewAccountDetail:
		content = m.viewAccountDetailPage()
	}

	return content
}

func (m Model) viewModal() string {
	switch m.modalType {
	case ModalAddTransaction:
		return m.transactionForm.View()
	case ModalDeleteConfirm, ModalAbandonConfirm:
		return m.confirmDialog.View()
	case ModalInsights:
		return m.viewInsightsModal()
	}
	return ""
}

func (m Model) viewInsightsModal() string {
	if m.insightsView == nil {
		return ""
	}

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		m.insightsView.View(),
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#1a1a2e")),
	)
}

func (m Model) viewLoading() string {
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	text := lipgloss.JoinVertical(lipgloss.Center,
		m.styles.Title.Render("Portfolio Tracker"),
		"",
		m.spinner.View()+" Loading portfolio data...",
	)

	return style.Render(text)
}

func (m Model) viewAccountsPage() string {
	header := m.header.View()
	body := m.accountsView.View()
	statusBar := m.statusBar.View()
	helpBar := m.styles.StatusBar.Width(m.width).Render(m.help.View(AccountsKeys))

	return lipgloss.JoinVertical(lipgloss.Left, header, body, statusBar, helpBar)
}

func (m Model) viewAccountDetailPage() string {
	header := m.header.View()
	body := m.accountDetailView.View()
	statusBar := m.statusBar.View()
	helpBar := m.styles.StatusBar.Width(m.width).Render(m.help.View(AccountDetailKeys))

	return lipgloss.JoinVertical(lipgloss.Left, header, body, statusBar, helpBar)
}

func collectUniqueTags(accounts *[]types.Account) []string {
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

func (m Model) getPortfolioInsights() tea.Cmd {
	return func() tea.Msg {
		client, err := llm.NewClient()
		if err != nil {
			return InsightsErrorMsg{
				Title: "Portfolio Insights",
				Error: err.Error(),
			}
		}

		filteredIds := m.getFilteredAccountIds()

		// Get transactions data from all filtered accounts
		var transactions []llm.TransactionData
		for _, id := range filteredIds {
			accountData := m.accountsData[id]
			for _, tx := range accountData.Transactions {
				transactions = append(transactions, llm.TransactionData{
					Date:     tx.Date.Format("2006-01-02"),
					Action:   string(tx.Type),
					Symbol:   tx.Symbol,
					Quantity: fmt.Sprintf("%.2f", float64(tx.Quantity)),
					Price:    fmt.Sprintf("%.2f", float64(tx.Pps)/100),
				})
			}
		}

		// Get metrics from the portfolio
		portfolioData := m.allPortfolio
		if m.tagFilter != "All" {
			portfolioData, _ = portfolio.LoadAndAnalyzeAccounts(m.db, filteredIds)
		}

		metrics := llm.MetricsData{
			Date:                  "",
			TotalValue:            fmt.Sprintf("%.2f", float64(portfolioData.Value)/100),
			CostBasis:             fmt.Sprintf("%.2f", float64(portfolioData.TotalInvested)/100),
			UnrealizedGain:        fmt.Sprintf("%.2f", float64(portfolioData.GainValue)/100),
			UnrealizedGainPercent: fmt.Sprintf("%.2f", float64(portfolioData.Gain)),
			DividendsReceived:     fmt.Sprintf("%.2f", float64(portfolioData.TotalDividends)/100),
			YieldOnCost:           fmt.Sprintf("%.2f", float64(portfolioData.AnnualizedYield)),
		}

		// Build prompt with empty holdings (would need more data structure)
		var holdings []llm.PortfolioData
		userInput := llm.ContextBuilder("Full Portfolio", holdings, transactions, metrics)

		// Get response from Responses API
		content, err := client.Generate(llm.SystemPrompt, userInput, 2000)
		if err != nil {
			return InsightsErrorMsg{
				Title: "Portfolio Insights",
				Error: err.Error(),
			}
		}

		return InsightsLoadedMsg{
			Title:   "Portfolio Insights",
			Content: content,
		}
	}
}

func (m Model) getAccountInsights() tea.Cmd {
	return func() tea.Msg {
		client, err := llm.NewClient()
		if err != nil {
			return InsightsErrorMsg{
				Title: "Account Insights",
				Error: err.Error(),
			}
		}

		accountData := m.accountsData[m.selectedAccount.Id]

		// Get transactions data
		var transactions []llm.TransactionData
		for _, tx := range accountData.Transactions {
			transactions = append(transactions, llm.TransactionData{
				Date:     tx.Date.Format("2006-01-02"),
				Action:   string(tx.Type),
				Symbol:   tx.Symbol,
				Quantity: fmt.Sprintf("%.2f", float64(tx.Quantity)),
				Price:    fmt.Sprintf("%.2f", float64(tx.Pps)/100),
			})
		}

		// Get metrics
		metrics := llm.MetricsData{
			Date:                  "",
			TotalValue:            fmt.Sprintf("%.2f", float64(accountData.Value)/100),
			CostBasis:             fmt.Sprintf("%.2f", float64(accountData.TotalInvested)/100),
			UnrealizedGain:        fmt.Sprintf("%.2f", float64(accountData.GainValue)/100),
			UnrealizedGainPercent: fmt.Sprintf("%.2f", float64(accountData.Gain)),
			DividendsReceived:     fmt.Sprintf("%.2f", float64(accountData.TotalDividends)/100),
			YieldOnCost:           fmt.Sprintf("%.2f", float64(accountData.AnnualizedYield)),
		}

		// Build prompt with empty holdings (would need more data structure)
		var holdings []llm.PortfolioData
		userInput := llm.ContextBuilder(m.selectedAccount.Name, holdings, transactions, metrics)

		// Get response from Responses API
		content, err := client.Generate(llm.SystemPrompt, userInput, 2000)
		if err != nil {
			return InsightsErrorMsg{
				Title: fmt.Sprintf("Account: %s Insights", m.selectedAccount.Name),
				Error: err.Error(),
			}
		}

		return InsightsLoadedMsg{
			Title:   fmt.Sprintf("Account: %s Insights", m.selectedAccount.Name),
			Content: content,
		}
	}
}
