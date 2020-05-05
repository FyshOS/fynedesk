package builtin

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
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

func (b *battery) value() (float64, error) {
	nowFile, fullFile := pickChargeOrEnergy()
	nowStr, err1 := ioutil.ReadFile(nowFile)
	fullStr, err2 := ioutil.ReadFile(fullFile)
	if err1 != nil || err2 != nil {
		log.Println("Error reading battery info", err1)
		return 0, err1
	}

	now, err1 := strconv.Atoi(strings.TrimSpace(string(nowStr)))
	full, err2 := strconv.Atoi(strings.TrimSpace(string(fullStr)))
	if err1 != nil || err2 != nil {
		log.Println("Error converting battery info", err1)
		return 0, err1
	}

	return float64(now) / float64(full), nil
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

func (b *battery) Shortcuts() map[*fynedesk.Shortcut]func() {
	return nil
}

func (b *battery) StatusAreaWidget() fyne.CanvasObject {
	if _, err := b.value(); err != nil {
		return nil
	}

	b.bar = widget.NewProgressBar()
	batteryIcon := widget.NewIcon(wmtheme.BatteryIcon)

	go b.batteryTick()
	return fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, batteryIcon, nil), batteryIcon, b.bar)
}

// newBattery creates a new module that will show battery level in the status area
func newBattery() fynedesk.Module {
	return &battery{}
}
