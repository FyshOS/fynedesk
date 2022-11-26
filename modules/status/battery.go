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
	fill *canvas.Rectangle
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
			<-tick.C
			val, _ := b.value()
			b.setValue(val)
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
	b.fill = canvas.NewRectangle(theme.ForegroundColor())
	prop := canvas.NewRectangle(color.Transparent)
	prop.SetMinSize(b.icon.MinSize().Add(fyne.NewSize(theme.Padding()*4, 0)))
	icon := container.NewMax(container.NewCenter(prop, b.icon), container.NewWithoutLayout(b.fill))

	// Set first value then tick
	val, _ := b.value()
	b.setValue(val)
	go b.batteryTick()
	return container.New(&handleNarrow{}, icon, b.bar)
}

func (b *battery) positionFill(val float64) {
	max := float32(12)
	down := max - max*float32(val)
	b.fill.Move(fyne.NewPos(14, 13+down))
	b.fill.Resize(fyne.NewSize(8, 13-down))
}

func (b *battery) setValue(val float64) {
	b.bar.SetValue(val)
	b.positionFill(val)
	if on, err := b.powered(); on || err != nil {
		b.icon.SetResource(wmtheme.PowerIcon)
		b.fill.Hide()
	} else if val < 0.1 {
		b.icon.SetResource(theme.NewErrorThemedResource(wmtheme.BatteryIcon))
		b.fill.FillColor = theme.ErrorColor()
		b.fill.Refresh()
		b.fill.Show()
	} else {
		b.icon.SetResource(wmtheme.BatteryIcon)
		b.fill.FillColor = theme.ForegroundColor()
		b.fill.Refresh()
		b.fill.Show()
	}
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

	if fynedesk.Instance() != nil && fynedesk.Instance().Settings().NarrowWidgetPanel() {
		objects[1].Hide()
	} else {
		objects[1].Show()
	}
}

func (h *handleNarrow) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(36, 36)
}
