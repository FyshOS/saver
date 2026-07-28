package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type darkTheme struct {
	fyne.Theme
}

func newDarkBackgroundTheme() *darkTheme {
	return &darkTheme{Theme: theme.DefaultTheme()}
}

func (t *darkTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameBackground:
		return color.Black
	}
	return t.Theme.Color(n, v)
}
