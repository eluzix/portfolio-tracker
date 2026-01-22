package views

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

type HelpView struct {
	width    int
	height   int
	bindings [][]key.Binding
	styles   HelpStyles
}

type HelpStyles struct {
	Overlay    lipgloss.Style
	Container  lipgloss.Style
	Title      lipgloss.Style
	Section    lipgloss.Style
	Key        lipgloss.Style
	Desc       lipgloss.Style
	Hint       lipgloss.Style
}

func DefaultHelpStyles() HelpStyles {
	return HelpStyles{
		Overlay: lipgloss.NewStyle().
			Background(lipgloss.Color("#1a1a2e")),
		Container: lipgloss.NewStyle().
			Background(lipgloss.Color("#27273a")).
			Foreground(lipgloss.Color("#eaeaea")).
			Padding(1, 3).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7c3aed")),
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7c3aed")).
			Bold(true).
			MarginBottom(1),
		Section: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a5b4fc")).
			Bold(true).
			MarginTop(1),
		Key: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#22d3ee")).
			Bold(true).
			Width(12),
		Desc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280")),
		Hint: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280")).
			Italic(true).
			MarginTop(1),
	}
}

func NewHelpView(bindings [][]key.Binding) HelpView {
	return HelpView{
		bindings: bindings,
		styles:   DefaultHelpStyles(),
	}
}

func (h *HelpView) SetSize(width, height int) {
	h.width = width
	h.height = height
}

func (h *HelpView) SetBindings(bindings [][]key.Binding) {
	h.bindings = bindings
}

func (h HelpView) View() string {
	sections := []string{
		h.styles.Title.Render("⌨  Keyboard Shortcuts"),
	}

	sectionNames := []string{"Navigation", "Currency & Filter", "Transactions", "General"}

	for i, group := range h.bindings {
		if i < len(sectionNames) {
			sections = append(sections, h.styles.Section.Render(sectionNames[i]))
		}

		for _, binding := range group {
			help := binding.Help()
			row := lipgloss.JoinHorizontal(lipgloss.Left,
				h.styles.Key.Render(help.Key),
				h.styles.Desc.Render(help.Desc),
			)
			sections = append(sections, row)
		}
	}

	sections = append(sections, h.styles.Hint.Render("Press ? to close"))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	modal := h.styles.Container.Render(content)

	return lipgloss.Place(h.width, h.height,
		lipgloss.Center, lipgloss.Center,
		modal,
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#1a1a2e")),
	)
}

func RenderHelpOverlay(width, height int, bindings [][]key.Binding) string {
	help := NewHelpView(bindings)
	help.SetSize(width, height)
	return help.View()
}
