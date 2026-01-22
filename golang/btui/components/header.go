package components

import (
	"github.com/charmbracelet/lipgloss"
)

type Header struct {
	width    int
	title    string
	subtitle string
	styles   HeaderStyles
}

type HeaderStyles struct {
	Bar      lipgloss.Style
	Title    lipgloss.Style
	Subtitle lipgloss.Style
}

func DefaultHeaderStyles() HeaderStyles {
	return HeaderStyles{
		Bar: lipgloss.NewStyle().
			Background(lipgloss.Color("#312e81")).
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 2),
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Bold(true),
		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a5b4fc")),
	}
}

func NewHeader() Header {
	return Header{
		title:  "Portfolio Tracker",
		styles: DefaultHeaderStyles(),
	}
}

func (h *Header) SetWidth(width int) {
	h.width = width
}

func (h *Header) SetTitle(title string) {
	h.title = title
}

func (h *Header) SetSubtitle(subtitle string) {
	h.subtitle = subtitle
}

func (h Header) View() string {
	titleSection := h.styles.Title.Render(h.title)

	var content string
	if h.subtitle != "" {
		subtitleSection := h.styles.Subtitle.Render(" │ " + h.subtitle)
		content = lipgloss.JoinHorizontal(lipgloss.Center, titleSection, subtitleSection)
	} else {
		content = titleSection
	}

	return h.styles.Bar.Width(h.width).Render(content)
}
