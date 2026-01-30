package components

import (
	"github.com/charmbracelet/lipgloss"
)

type Modal struct {
	width   int
	height  int
	content string
	visible bool
	styles  ModalStyles
}

type ModalStyles struct {
	Overlay   lipgloss.Style
	Container lipgloss.Style
}

func DefaultModalStyles() ModalStyles {
	return ModalStyles{
		Overlay: lipgloss.NewStyle().
			Background(lipgloss.Color("#1a1a2e")),
		Container: lipgloss.NewStyle().
			Background(lipgloss.Color("#27273a")).
			Foreground(lipgloss.Color("#eaeaea")).
			Padding(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7c3aed")),
	}
}

func NewModal() Modal {
	return Modal{
		styles: DefaultModalStyles(),
	}
}

func (m *Modal) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Modal) SetContent(content string) {
	m.content = content
}

func (m *Modal) Show() {
	m.visible = true
}

func (m *Modal) Hide() {
	m.visible = false
}

func (m *Modal) Visible() bool {
	return m.visible
}

func (m Modal) View() string {
	if !m.visible {
		return ""
	}

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		m.styles.Container.Render(m.content),
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#1a1a2e")),
	)
}

func RenderModalOverlay(width, height int, content string) string {
	return lipgloss.Place(width, height,
		lipgloss.Center, lipgloss.Center,
		content,
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#1a1a2e")),
	)
}
