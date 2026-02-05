package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type InsightsView struct {
	viewport viewport.Model
	title    string
	content  string
	styles   InsightsStyles
	ready    bool
}

type InsightsStyles struct {
	Overlay   lipgloss.Style
	Container lipgloss.Style
	Header    lipgloss.Style
	Footer    lipgloss.Style
	Content   lipgloss.Style
}

func DefaultInsightsStyles() InsightsStyles {
	return InsightsStyles{
		Overlay: lipgloss.NewStyle().
			Background(lipgloss.Color("#1a1a2e")).
			Foreground(lipgloss.Color("#ffffff")),
		Container: lipgloss.NewStyle().
			Background(lipgloss.Color("#24283b")).
			Foreground(lipgloss.Color("#e0e6f0")).
			Padding(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7aa2f7")),
		Header: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7dcfff")).
			Bold(true).
			MarginBottom(1),
		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#737aa2")).
			Italic(true).
			MarginTop(1),
		Content: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e0e6f0")),
	}
}

func NewInsightsView(title, content string) InsightsView {
	vp := viewport.New(0, 0)
	vp.YPosition = 0
	vp.HighPerformanceRendering = false

	iv := InsightsView{
		viewport: vp,
		title:    title,
		content:  content,
		styles:   DefaultInsightsStyles(),
		ready:    false,
	}

	return iv
}

func (i *InsightsView) SetSize(width, height int) {
	// Modal dimensions: 80% width, 70% height, centered
	modalWidth := (width * 80) / 100
	modalHeight := (height * 70) / 100

	if modalWidth < 40 {
		modalWidth = 40
	}
	if modalHeight < 10 {
		modalHeight = 10
	}

	// Account for borders and padding (2 chars for borders, 4 for padding)
	contentWidth := modalWidth - 4
	contentHeight := modalHeight - 4 // Account for header, footer, and borders

	if contentWidth < 20 {
		contentWidth = 20
	}
	if contentHeight < 5 {
		contentHeight = 5
	}

	i.viewport.Width = contentWidth
	i.viewport.Height = contentHeight
	i.viewport.SetContent(i.content)
	i.ready = true
}

func (i *InsightsView) ScrollUp() {
	if i.viewport.YOffset > 0 {
		i.viewport.YOffset--
	}
}

func (i *InsightsView) ScrollDown() {
	maxScroll := len(strings.Split(i.content, "\n")) - i.viewport.Height
	if i.viewport.YOffset < maxScroll {
		i.viewport.YOffset++
	}
}

func (i *InsightsView) SetContent(title, content string) {
	i.title = title
	i.content = content
	i.viewport.SetContent(content)
	i.viewport.YOffset = 0
}

func (i InsightsView) View() string {
	if !i.ready {
		return "Loading..."
	}

	totalLines := len(strings.Split(i.content, "\n"))
	scrollPos := i.viewport.YOffset

	header := i.styles.Header.Render("📊 " + i.title)
	footer := i.styles.Footer.Render("Scroll: ↑/k/↓/j | Copy: c | Close: Esc | " +
		strings.TrimSpace(strings.TrimRight(strings.TrimLeft(
			i.styles.Footer.Render(strings.Join([]string{
				"[" + renderScrollPosition(scrollPos, totalLines) + "]",
			}, "")), " "), " ")))

	content := i.styles.Content.Render(i.viewport.View())

	modalContent := lipgloss.JoinVertical(lipgloss.Top,
		header,
		content,
		footer,
	)

	modal := i.styles.Container.Render(modalContent)

	return modal
}

func renderScrollPosition(current, total int) string {
	percent := 0
	if total > 0 {
		percent = (current * 100) / total
	}
	return lipgloss.NewStyle().Render(strings.TrimSpace(lipgloss.NewStyle().Render(
		strings.Join([]string{"Line " + string(rune(current+1)) + "/" + string(rune(total)) + " (" + string(rune(rune(percent))) + "%)"}, ""))))
}

func RenderInsightsModal(width, height int, title, content string) string {
	modal := NewInsightsView(title, content)
	modal.SetSize(width, height)

	// Calculate modal dimensions
	modalWidth := (width * 80) / 100
	modalHeight := (height * 70) / 100
	if modalWidth < 40 {
		modalWidth = 40
	}
	if modalHeight < 10 {
		modalHeight = 10
	}

	return lipgloss.Place(width, height,
		lipgloss.Center, lipgloss.Center,
		modal.View(),
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#1a1a2e")),
	)
}
