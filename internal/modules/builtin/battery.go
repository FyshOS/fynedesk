package builtin

import (
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"

	"fyne.io/desktop"
	wmtheme "fyne.io/desktop/theme"
)

type battery struct {
	bar *widget.ProgressBar
}

func (b *battery) value() (float64, error) {
	nowStr, err1 := ioutil.ReadFile("/sys/class/power_supply/BAT0/charge_now")
	fullStr, err2 := ioutil.ReadFile("/sys/class/power_supply/BAT0/charge_full")
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
		for {
			val, _ := b.value()
			b.bar.SetValue(val)
			<-tick.C
		}
	}()
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

func (b *battery) Metadata() desktop.ModuleMetadata {
	return desktop.ModuleMetadata{
		Name: "Battery",
	}
}

// NewBattery creates a new module that will show battery level in the status area
func NewBattery() desktop.Module {
	return &battery{}
}
