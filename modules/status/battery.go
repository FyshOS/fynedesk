package status

import (
	"image/color"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

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
	icon *widget.Icon
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
			if on, err := b.powered(); on || err != nil {
				b.icon.SetResource(wmtheme.PowerIcon)
			} else if val < 0.1 {
				b.icon.SetResource(theme.NewErrorThemedResource(wmtheme.BatteryIcon))
			} else {
				b.icon.SetResource(wmtheme.BatteryIcon)
			}
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
	b.icon = widget.NewIcon(wmtheme.BatteryIcon)
	prop := canvas.NewRectangle(color.Transparent)
	prop.SetMinSize(b.icon.MinSize().Add(fyne.NewSize(theme.Padding()*4, 0)))
	icon := container.NewCenter(prop, b.icon)

	go b.batteryTick()
	return container.New(&handleNarrow{}, icon, b.bar)
}

// newBattery creates a new module that will show battery level in the status area
func newBattery() fynedesk.Module {
	return &battery{}
}

type handleNarrow struct{}

func (h *handleNarrow) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	objects[0].Resize(fyne.NewSize(size.Height, size.Height))
	objects[1].Resize(fyne.NewSize(size.Width-size.Height-theme.Padding(), size.Height))
	objects[1].Move(fyne.NewPos(size.Height+theme.Padding(), 0))

	if fynedesk.Instance().Settings().NarrowWidgetPanel() {
		objects[1].Hide()
	} else {
		objects[1].Show()
	}
}

func (h *handleNarrow) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(36, 36)
}
