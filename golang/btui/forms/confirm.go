package forms

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ConfirmDialog struct {
	title     string
	message   string
	width     int
	height    int
	confirmed bool
	completed bool
	focused   int
	styles    ConfirmDialogStyles
}

type ConfirmDialogStyles struct {
	Container    lipgloss.Style
	Title        lipgloss.Style
	Message      lipgloss.Style
	Button       lipgloss.Style
	ButtonActive lipgloss.Style
	ButtonDanger lipgloss.Style
}

func DefaultConfirmDialogStyles() ConfirmDialogStyles {
	return ConfirmDialogStyles{
		Container: lipgloss.NewStyle().
			Background(lipgloss.Color("#27273a")).
			Foreground(lipgloss.Color("#eaeaea")).
			Padding(1, 3).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7c3aed")).
			Width(50),
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ef4444")).
			Bold(true).
			MarginBottom(1),
		Message: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#eaeaea")).
			MarginBottom(1),
		Button: lipgloss.NewStyle().
			Background(lipgloss.Color("#475569")).
			Foreground(lipgloss.Color("#eaeaea")).
			Padding(0, 2).
			MarginRight(2),
		ButtonActive: lipgloss.NewStyle().
			Background(lipgloss.Color("#6366f1")).
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 2).
			MarginRight(2).
			Bold(true),
		ButtonDanger: lipgloss.NewStyle().
			Background(lipgloss.Color("#ef4444")).
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 2).
			MarginRight(2).
			Bold(true),
	}
}

func NewConfirmDialog(title, message string) ConfirmDialog {
	return ConfirmDialog{
		title:   title,
		message: message,
		focused: 1,
		styles:  DefaultConfirmDialogStyles(),
	}
}

func NewDeleteConfirmDialog(message string) ConfirmDialog {
	return ConfirmDialog{
		title:   "Delete Transaction?",
		message: message,
		focused: 1,
		styles:  DefaultConfirmDialogStyles(),
	}
}

func NewAbandonConfirmDialog() ConfirmDialog {
	return ConfirmDialog{
		title:   "Abandon Changes?",
		message: "Are you sure you want to abandon your changes?",
		focused: 1,
		styles:  DefaultConfirmDialogStyles(),
	}
}

func (d *ConfirmDialog) SetSize(width, height int) {
	d.width = width
	d.height = height
}

func (d *ConfirmDialog) Completed() bool {
	return d.completed
}

func (d *ConfirmDialog) Confirmed() bool {
	return d.confirmed
}

func (d *ConfirmDialog) Update(msg tea.Msg) (ConfirmDialog, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, key.NewBinding(key.WithKeys("left", "h"))):
			d.focused = 0
		case key.Matches(keyMsg, key.NewBinding(key.WithKeys("right", "l"))):
			d.focused = 1
		case key.Matches(keyMsg, key.NewBinding(key.WithKeys("tab"))):
			d.focused = (d.focused + 1) % 2
		case key.Matches(keyMsg, key.NewBinding(key.WithKeys("enter"))):
			d.completed = true
			d.confirmed = d.focused == 0
		case key.Matches(keyMsg, key.NewBinding(key.WithKeys("y"))):
			d.completed = true
			d.confirmed = true
		case key.Matches(keyMsg, key.NewBinding(key.WithKeys("n", "esc"))):
			d.completed = true
			d.confirmed = false
		}
	}

	return *d, nil
}

func (d ConfirmDialog) View() string {
	title := d.styles.Title.Render(d.title)
	message := d.styles.Message.Render(d.message)

	var yesButton, noButton string
	if d.focused == 0 {
		yesButton = d.styles.ButtonDanger.Render("Yes, confirm")
		noButton = d.styles.Button.Render("No, cancel")
	} else {
		yesButton = d.styles.Button.Render("Yes, confirm")
		noButton = d.styles.ButtonActive.Render("No, cancel")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Left, yesButton, noButton)
	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6b7280")).
		Italic(true).
		MarginTop(1).
		Render("y/n to choose • ←/→ to navigate • enter to confirm")

	content := lipgloss.JoinVertical(lipgloss.Left, title, message, buttons, hint)
	modal := d.styles.Container.Render(content)

	return lipgloss.Place(d.width, d.height,
		lipgloss.Center, lipgloss.Center,
		modal,
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#1a1a2e")),
	)
}
