package builtin

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	"fyne.io/fynedesk"
	wmtheme "fyne.io/fynedesk/theme"
)

var brightnessMeta = fynedesk.ModuleMetadata{
	Name:        "Brightness",
	NewInstance: newBrightness,
}

// Brightness is a progress bar module to modify screen brightness
type brightness struct {
	bar *widget.ProgressBar
}

func (b *brightness) Destroy() {
}

func (b *brightness) value() (float64, error) {
	out, err := exec.Command("xbacklight").Output()
	if err != nil {
		log.Println("Error running xbacklight", err)
		return 0, err
	}
	ret, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		log.Println("Error reading brightness info", err)
		return 0, err
	}
	return ret / 100, nil
}

func (b *brightness) offsetValue(diff int) {
	floatVal, _ := b.value()
	value := int(floatVal*100) + diff

	if value < 5 {
		value = 5
	} else if value > 100 {
		value = 100
	}

	err := exec.Command("xbacklight", "-set", fmt.Sprintf("%d", value)).Run()
	if err != nil {
		log.Println("Error running xbacklight", err)
	} else {
		newVal, _ := b.value()
		b.bar.SetValue(newVal)
	}
}

func (b *brightness) Metadata() fynedesk.ModuleMetadata {
	return brightnessMeta
}

func (b *brightness) Shortcuts() map[*fynedesk.Shortcut]func() {
	return map[*fynedesk.Shortcut]func(){
		fynedesk.NewShortcut("Increase Screen Brightness", fynedesk.KeyBrightnessDown, fynedesk.AnyModifier): func() {
			b.offsetValue(-5)
		},
		fynedesk.NewShortcut("Reduce Screen Brightness", fynedesk.KeyBrightnessUp, fynedesk.AnyModifier): func() {
			b.offsetValue(5)
		},
	}
}

func (b *brightness) StatusAreaWidget() fyne.CanvasObject {
	if _, err := b.value(); err != nil {
		return nil
	}

	b.bar = widget.NewProgressBar()
	brightnessIcon := widget.NewIcon(wmtheme.BrightnessIcon)
	less := widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() {
		b.offsetValue(-5)
	})
	more := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		b.offsetValue(5)
	})
	bright := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, less, more),
		less, b.bar, more)

	go b.offsetValue(0)
	return fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, brightnessIcon, nil), brightnessIcon, bright)
}

// newBrightness creates a new module that will show screen brightness in the status area
func newBrightness() fynedesk.Module {
	return &brightness{}
}
