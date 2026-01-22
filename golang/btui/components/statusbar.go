package components

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type StatusBar struct {
	width      int
	mode       string
	hint       string
	status     string
	loading    bool
	spinner    spinner.Model
	styles     StatusBarStyles
}

type StatusBarStyles struct {
	Bar      lipgloss.Style
	Mode     lipgloss.Style
	Hint     lipgloss.Style
	Status   lipgloss.Style
	Spinner  lipgloss.Style
}

func DefaultStatusBarStyles() StatusBarStyles {
	return StatusBarStyles{
		Bar: lipgloss.NewStyle().
			Background(lipgloss.Color("#312e81")).
			Foreground(lipgloss.Color("#eaeaea")).
			Padding(0, 1),
		Mode: lipgloss.NewStyle().
			Background(lipgloss.Color("#7c3aed")).
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 1).
			Bold(true),
		Hint: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280")),
		Status: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#22d3ee")),
		Spinner: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#22d3ee")),
	}
}

func NewStatusBar() StatusBar {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#22d3ee"))

	return StatusBar{
		mode:    "NORMAL",
		styles:  DefaultStatusBarStyles(),
		spinner: s,
	}
}

func (s *StatusBar) SetWidth(width int) {
	s.width = width
}

func (s *StatusBar) SetMode(mode string) {
	s.mode = mode
}

func (s *StatusBar) SetHint(hint string) {
	s.hint = hint
}

func (s *StatusBar) SetStatus(status string) {
	s.status = status
}

func (s *StatusBar) SetLoading(loading bool) {
	s.loading = loading
}

func (s *StatusBar) UpdateSpinner(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	s.spinner, cmd = s.spinner.Update(msg)
	return cmd
}

func (s StatusBar) View() string {
	modeSection := s.styles.Mode.Render(s.mode)

	var statusSection string
	if s.loading {
		statusSection = s.styles.Spinner.Render(s.spinner.View()) + " " + s.styles.Status.Render(s.status)
	} else if s.status != "" {
		statusSection = s.styles.Status.Render(s.status)
	}

	hintSection := s.styles.Hint.Render(s.hint)

	modeWidth := lipgloss.Width(modeSection)
	statusWidth := lipgloss.Width(statusSection)
	hintWidth := lipgloss.Width(hintSection)

	gap := s.width - modeWidth - statusWidth - hintWidth - 4
	if gap < 0 {
		gap = 1
	}

	content := lipgloss.JoinHorizontal(lipgloss.Center,
		modeSection,
		"  ",
		statusSection,
		lipgloss.NewStyle().Width(gap).Render(""),
		hintSection,
	)

	return s.styles.Bar.Width(s.width).Render(content)
}
