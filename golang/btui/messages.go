package btui

import "tracker/types"

type DataLoadedMsg struct {
	Accounts     *[]types.Account
	AccountsData map[string]types.AnalyzedPortfolio
	AllPortfolio types.AnalyzedPortfolio
}

type AccountSelectedMsg struct {
	Account types.Account
}

type BackToAccountsMsg struct{}

type CurrencyChangedMsg struct {
	Symbol       string
	ExchangeRate float64
}

type TagFilterChangedMsg struct {
	Tag string
}

type TransactionAddedMsg struct {
	Transaction types.Transaction
}

type TransactionDeletedMsg struct {
	TransactionID string
}

type ToggleDividendsMsg struct{}

type ShowHelpMsg struct{}

type HideHelpMsg struct{}

type ShowModalMsg struct {
	Type    ModalType
	Payload any
}

type HideModalMsg struct{}

type ModalType int

const (
	ModalNone ModalType = iota
	ModalAddTransaction
	ModalDeleteConfirm
	ModalAbandonConfirm
)

type ErrorMsg struct {
	Err error
}

func (e ErrorMsg) Error() string {
	return e.Err.Error()
}

type StatusMsg struct {
	Text string
}

type ClearStatusMsg struct{}
