package tui

import "github.com/gdamore/tcell/v2"

// Theme defines the color scheme for the application
type Theme struct {
	Background     tcell.Color
	Foreground     tcell.Color
	Border         tcell.Color
	HeaderBg       tcell.Color
	HeaderFg       tcell.Color
	ButtonBg       tcell.Color
	ButtonFg       tcell.Color
	Positive       tcell.Color
	Negative       tcell.Color
	ModalBg        tcell.Color
	ModalFg        tcell.Color
	SelectedBg     tcell.Color
	SelectedFg     tcell.Color
}

// DefaultTheme returns a slick dark theme
func DefaultTheme() Theme {
	return Theme{
		Background:   tcell.ColorBlack,
		Foreground:   tcell.ColorWhite,
		Border:       tcell.ColorLightBlue,
		HeaderBg:     tcell.ColorBlue,
		HeaderFg:     tcell.ColorWhite,
		ButtonBg:     tcell.ColorGold,
		ButtonFg:     tcell.ColorBlack,
		Positive:     tcell.ColorLime,
		Negative:     tcell.ColorRed,
		ModalBg:      tcell.ColorDarkGray,
		ModalFg:      tcell.ColorWhite,
		SelectedBg:   tcell.ColorOlive,
		SelectedFg:   tcell.ColorBlack,
	}
}

// GetTheme returns the current theme (for now, default)
var GetTheme = DefaultTheme
