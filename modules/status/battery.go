package status

import (
	"image/color"
	"os"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	"fyne.io/fynedesk"
	wmtheme "fyne.io/fynedesk/theme"
)

var batteryMeta = fynedesk.ModuleMetadata{
	Name:        "Battery",
	NewInstance: newBattery,
}

type battery struct {
	bar  *widget.ProgressBar
	done bool
}

func pickChargeOrEnergy() (string, string) {
	_, err := os.Stat("/sys/class/power_supply/BAT0/charge_now")
	if err != nil {
		return "/sys/class/power_supply/BAT0/energy_now", "/sys/class/power_supply/BAT0/energy_full"
	}
	return "/sys/class/power_supply/BAT0/charge_now", "/sys/class/power_supply/BAT0/charge_full"
}

func (b *battery) batteryTick() {
	tick := time.NewTicker(time.Second * 10)
	go func() {
		for !b.done {
			val, _ := b.value()
			b.bar.SetValue(val)
			<-tick.C
		}
	}()
}

func (b *battery) Destroy() {
	b.done = true
}

func (b *battery) Metadata() fynedesk.ModuleMetadata {
	return batteryMeta
}

func (b *battery) StatusAreaWidget() fyne.CanvasObject {
	if _, err := b.value(); err != nil {
		return nil
	}

	b.bar = widget.NewProgressBar()
	batteryIcon := widget.NewIcon(wmtheme.BatteryIcon)
	prop := canvas.NewRectangle(color.Transparent)
	prop.SetMinSize(batteryIcon.MinSize().Add(fyne.NewSize(theme.Padding()*2, 0)))
	icon := container.NewCenter(prop, batteryIcon)

	go b.batteryTick()
	return container.NewBorder(nil, nil, icon, nil, b.bar)
}

// newBattery creates a new module that will show battery level in the status area
func newBattery() fynedesk.Module {
	return &battery{}
}
