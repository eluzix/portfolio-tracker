package btui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Background   lipgloss.Color
	Foreground   lipgloss.Color
	Primary      lipgloss.Color
	Secondary    lipgloss.Color
	HeaderBg     lipgloss.Color
	HeaderFg     lipgloss.Color
	SelectedBg   lipgloss.Color
	SelectedFg   lipgloss.Color
	Positive     lipgloss.Color
	Negative     lipgloss.Color
	Muted        lipgloss.Color
	Border       lipgloss.Color
	ButtonBg     lipgloss.Color
	ButtonFg     lipgloss.Color
	ModalBg      lipgloss.Color
	ModalFg      lipgloss.Color
}

var DefaultTheme = Theme{
	Background:   lipgloss.Color("#1a1b26"),
	Foreground:   lipgloss.Color("#e0e6f0"),
	Primary:      lipgloss.Color("#bb9af7"),
	Secondary:    lipgloss.Color("#7dcfff"),
	HeaderBg:     lipgloss.Color("#24283b"),
	HeaderFg:     lipgloss.Color("#e0e6f0"),
	SelectedBg:   lipgloss.Color("#d4a5c8"),
	SelectedFg:   lipgloss.Color("#1a1b26"),
	Positive:     lipgloss.Color("#9ece6a"),
	Negative:     lipgloss.Color("#f7768e"),
	Muted:        lipgloss.Color("#737aa2"),
	Border:       lipgloss.Color("#7aa2f7"),
	ButtonBg:     lipgloss.Color("#e0af68"),
	ButtonFg:     lipgloss.Color("#1a1b26"),
	ModalBg:      lipgloss.Color("#24283b"),
	ModalFg:      lipgloss.Color("#e0e6f0"),
}

type Styles struct {
	App           lipgloss.Style
	Header        lipgloss.Style
	Title         lipgloss.Style
	StatusBar     lipgloss.Style
	StatusText    lipgloss.Style
	StatusSpinner lipgloss.Style
	Table         lipgloss.Style
	TableHeader   lipgloss.Style
	TableRow      lipgloss.Style
	TableSelected lipgloss.Style
	Button        lipgloss.Style
	ButtonActive  lipgloss.Style
	Help          lipgloss.Style
	HelpKey       lipgloss.Style
	HelpDesc      lipgloss.Style
	Modal         lipgloss.Style
	ModalTitle    lipgloss.Style
	Positive      lipgloss.Style
	Negative      lipgloss.Style
	Muted         lipgloss.Style
	Border        lipgloss.Style
}

func NewStyles(t Theme) Styles {
	return Styles{
		App: lipgloss.NewStyle().
			Background(t.Background).
			Foreground(t.Foreground),

		Header: lipgloss.NewStyle().
			Background(t.HeaderBg).
			Foreground(t.HeaderFg).
			Padding(0, 2).
			Bold(true),

		Title: lipgloss.NewStyle().
			Foreground(t.Primary).
			Bold(true),

		StatusBar: lipgloss.NewStyle().
			Background(t.HeaderBg).
			Foreground(t.HeaderFg).
			Padding(0, 1),

		StatusText: lipgloss.NewStyle().
			Foreground(t.Muted),

		StatusSpinner: lipgloss.NewStyle().
			Foreground(t.Secondary),

		Table: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(t.Border).
			Padding(1, 2),

		TableHeader: lipgloss.NewStyle().
			Background(t.HeaderBg).
			Foreground(t.HeaderFg).
			Bold(true).
			Padding(0, 1),

		TableRow: lipgloss.NewStyle().
			Foreground(t.Foreground).
			Padding(0, 1),

		TableSelected: lipgloss.NewStyle().
			Background(t.SelectedBg).
			Foreground(t.SelectedFg).
			Bold(true).
			Padding(0, 1),

		Button: lipgloss.NewStyle().
			Background(t.ButtonBg).
			Foreground(t.ButtonFg).
			Padding(0, 2).
			MarginRight(1),

		ButtonActive: lipgloss.NewStyle().
			Background(t.Primary).
			Foreground(t.HeaderFg).
			Padding(0, 2).
			MarginRight(1).
			Bold(true),

		Help: lipgloss.NewStyle().
			Background(t.ModalBg).
			Foreground(t.ModalFg).
			Padding(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(t.Border),

		HelpKey: lipgloss.NewStyle().
			Foreground(t.Secondary).
			Bold(true),

		HelpDesc: lipgloss.NewStyle().
			Foreground(t.Muted),

		Modal: lipgloss.NewStyle().
			Background(t.ModalBg).
			Foreground(t.ModalFg).
			Padding(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(t.Primary),

		ModalTitle: lipgloss.NewStyle().
			Foreground(t.Primary).
			Bold(true).
			MarginBottom(1),

		Positive: lipgloss.NewStyle().
			Foreground(t.Positive),

		Negative: lipgloss.NewStyle().
			Foreground(t.Negative),

		Muted: lipgloss.NewStyle().
			Foreground(t.Muted),

		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(t.Border),
	}
}

var AppStyles = NewStyles(DefaultTheme)
