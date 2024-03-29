package status

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestBattery_Render(t *testing.T) {
	b := newBattery().(*battery)
	wid := b.StatusAreaWidget()
	if wid == nil { // we don't have a test stub value
		return
	}
	w := test.NewWindow(wid)
	w.Resize(fyne.NewSize(103, 44))

	b.setValue(1)
	test.AssertImageMatches(t, "battery_full.png", w.Canvas().Capture())

	b.setValue(0.5)
	test.AssertImageMatches(t, "battery_50.png", w.Canvas().Capture())

	b.setValue(0.25)
	test.AssertImageMatches(t, "battery_25.png", w.Canvas().Capture())
}

func TestBattery_Render_LowWarning(t *testing.T) {
	b := newBattery().(*battery)
	wid := b.StatusAreaWidget()
	if wid == nil { // we don't have a test stub value
		return
	}
	w := test.NewWindow(wid)
	w.Resize(fyne.NewSize(103, 44))

	b.setValue(0.09)
	test.AssertImageMatches(t, "battery_low.png", w.Canvas().Capture())
}
