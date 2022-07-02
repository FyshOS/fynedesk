package status

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

func TestBattery_Render(t *testing.T) {
	test.NewApp().Settings().SetTheme(theme.DarkTheme())
	b := newBattery().(*battery)
	wid := b.StatusAreaWidget()
	w := test.NewWindow(wid)

	b.setValue(1)
	test.AssertImageMatches(t, "battery_full.png", w.Canvas().Capture())

	b.setValue(0.5)
	test.AssertImageMatches(t, "battery_50.png", w.Canvas().Capture())

	b.setValue(0.25)
	test.AssertImageMatches(t, "battery_25.png", w.Canvas().Capture())
}

func TestBattery_Render_LowWarning(t *testing.T) {
	test.NewApp().Settings().SetTheme(theme.DarkTheme())
	b := newBattery().(*battery)
	wid := b.StatusAreaWidget()
	w := test.NewWindow(wid)

	b.setValue(0.09)
	test.AssertImageMatches(t, "battery_low.png", w.Canvas().Capture())
}
