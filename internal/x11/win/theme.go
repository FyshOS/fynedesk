//go:build linux || openbsd || freebsd || netbsd

package win

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type transparentTheme struct {
	fyne.Theme

	frame *frame
}

func (t *transparentTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameShadow:
		return color.Transparent
	case theme.ColorNameBackground, theme.ColorNameOverlayBackground:
		if t.frame.client.Focused() {
			n = theme.ColorNameOverlayBackground
		} else {
			n = theme.ColorNameDisabledButton
		}
	}

	return t.Theme.Color(n, v)
}
