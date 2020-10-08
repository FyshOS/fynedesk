package status

import (
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
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

func (b *battery) value() (float64, error) {
	if runtime.GOOS == "linux" {
		return b.valueLinux()
	}

	return b.valueBSD()
}

func (b *battery) valueBSD() (float64, error) {
	val, err := syscall.Sysctl("hw.acpi.battery.life")
	if err != nil {
		return 0, err
	}

	percent, err := strconv.Atoi(strings.TrimSpace(val))
	if err != nil || percent == 0 {
		return 0, err
	}

	return float64(percent)/100, nil
}

func (b *battery) valueLinux() (float64, error) {
	nowFile, fullFile := pickChargeOrEnergy()
	fullStr, err1 := ioutil.ReadFile(fullFile)
	if os.IsNotExist(err1) {
		return 0, err1 // return quietly if the file was not present (desktop?)
	}
	nowStr, err2 := ioutil.ReadFile(nowFile)
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

func (b *battery) StatusAreaWidget() fyne.CanvasObject {
	if _, err := b.value(); err != nil {
		return nil
	}

	b.bar = widget.NewProgressBar()
	batteryIcon := widget.NewIcon(wmtheme.BatteryIcon)
	prop := canvas.NewRectangle(color.Transparent)
	prop.SetMinSize(batteryIcon.MinSize().Add(fyne.NewSize(theme.Padding()*2, 0)))
	icon := fyne.NewContainerWithLayout(layout.NewCenterLayout(), prop, batteryIcon)

	go b.batteryTick()
	return fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, icon, nil), icon, b.bar)
}

// newBattery creates a new module that will show battery level in the status area
func newBattery() fynedesk.Module {
	return &battery{}
}
