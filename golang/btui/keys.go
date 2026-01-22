package btui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up            key.Binding
	Down          key.Binding
	Enter         key.Binding
	Back          key.Binding
	Quit          key.Binding
	Help          key.Binding
	Tab           key.Binding
	NewTx         key.Binding
	DeleteTx      key.Binding
	ToggleDivs    key.Binding
	CurrencyUSD   key.Binding
	CurrencyNIS   key.Binding
	CycleTag      key.Binding
	Confirm       key.Binding
	Cancel        key.Binding
}

var Keys = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "cycle focus"),
	),
	NewTx: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new transaction"),
	),
	DeleteTx: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete transaction"),
	),
	ToggleDivs: key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("h", "toggle dividends"),
	),
	CurrencyUSD: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "USD"),
	),
	CurrencyNIS: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "NIS"),
	),
	CycleTag: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "cycle tag"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "confirm"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("n", "esc"),
		key.WithHelp("n/esc", "cancel"),
	),
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Back, k.Help, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Back},
		{k.CurrencyUSD, k.CurrencyNIS, k.CycleTag},
		{k.NewTx, k.DeleteTx, k.ToggleDivs},
		{k.Tab, k.Help, k.Quit},
	}
}

type AccountsKeyMap struct {
	KeyMap
}

func (k AccountsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.CycleTag, k.Help, k.Quit}
}

func (k AccountsKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.CurrencyUSD, k.CurrencyNIS, k.CycleTag},
		{k.Help, k.Quit},
	}
}

type AccountDetailKeyMap struct {
	KeyMap
}

func (k AccountDetailKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.NewTx, k.DeleteTx, k.Back, k.Help}
}

func (k AccountDetailKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.NewTx, k.DeleteTx, k.ToggleDivs},
		{k.Back, k.Help, k.Quit},
	}
}

var AccountsKeys = AccountsKeyMap{Keys}
var AccountDetailKeys = AccountDetailKeyMap{Keys}
