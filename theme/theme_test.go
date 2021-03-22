package theme

import (
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
	fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme())
	battDark := BatteryIcon.Content()

	fyne.CurrentApp().Settings().SetTheme(theme.LightTheme())
	assert.NotEqual(t, battDark, BatteryIcon.Content())
}
