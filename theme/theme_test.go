package theme

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	_ "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"

	"github.com/stretchr/testify/assert"
)

func TestIconResources(t *testing.T) {
	assert.NotNil(t, BatteryIcon.Name())
	assert.NotNil(t, BrightnessIcon.Name())
	assert.NotNil(t, SoundIcon.Name())
	assert.NotNil(t, MuteIcon.Name())
}

func TestIconTheme(t *testing.T) {
	th := &testTheme{fg: color.White}
	fyne.CurrentApp().Settings().SetTheme(th)
	battDark := BatteryIcon.Content()

	th.fg = color.Black
	assert.NotEqual(t, battDark, BatteryIcon.Content())
}

func TestIconTheme_BrokenImage(t *testing.T) {
	assert.NotNil(t, BrokenImageIcon) // must not be nil as we fall back
}

type testTheme struct {
	fyne.Theme

	fg color.Color
}

func (t *testTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if n == theme.ColorNameForeground {
		return t.fg
	}

	return t.Theme.Color(n, v)
}
