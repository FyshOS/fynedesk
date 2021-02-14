package status

import (
	"errors"
	"fmt"
	"image/color"
	"os/exec"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

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
		fyne.LogError("Error running xbacklight", err)
		return 0, err
	}
	if strings.TrimSpace(string(out)) == "" {
		return 0, errors.New("no back-lit screens found")
	}
	ret, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		fyne.LogError("Error reading brightness info", err)
		return 0, err
	}
	return ret / 100, nil
}

func (b *brightness) offsetValue(diff int) {
	floatVal, _ := b.value()
	value := int(floatVal*100) + diff

	b.setValue(value)
}

func (b *brightness) setValue(value int) {
	if value < 5 {
		value = 5
	} else if value > 100 {
		value = 100
	}

	err := exec.Command("xbacklight", "-set", fmt.Sprintf("%d", value)).Run()
	if err != nil {
		fyne.LogError("Error running xbacklight", err)
	} else {
		newVal, _ := b.value()
		b.bar.SetValue(newVal)
	}
}

func (b *brightness) LaunchSuggestions(input string) []fynedesk.LaunchSuggestion {
	lower := strings.ToLower(input)
	matches := false
	val := lower
	if startsWith(lower, "brightness ") {
		matches = true
		if len(lower) > 11 {
			val = lower[11:]
		} else {
			val = ""
		}
	} else if startsWith(lower, "bright ") {
		matches = true
		if len(lower) > 7 {
			val = lower[7:]
		} else {
			val = ""
		}
	} else if startsWith(lower, "backlight ") {
		matches = true
		if len(lower) > 10 {
			val = lower[10:]
		} else {
			val = ""
		}
	}

	if matches {
		return []fynedesk.LaunchSuggestion{&brightItem{input: val, b: b}}
	}

	return nil
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
	prop := canvas.NewRectangle(color.Transparent)
	prop.SetMinSize(brightnessIcon.MinSize().Add(fyne.NewSize(theme.Padding()*2, 0)))
	icon := container.NewCenter(prop, brightnessIcon)

	less := &widget.Button{Icon: theme.ContentRemoveIcon(), Importance: widget.LowImportance, OnTapped: func() {
		b.offsetValue(-5)
	}}

	more := &widget.Button{Icon: theme.ContentAddIcon(), Importance: widget.LowImportance, OnTapped: func() {
		b.offsetValue(5)
	}}

	bright := container.NewBorder(nil, nil, less, more, b.bar)

	go b.offsetValue(0)
	return container.NewBorder(nil, nil, icon, nil, bright)
}

// newBrightness creates a new module that will show screen brightness in the status area
func newBrightness() fynedesk.Module {
	return &brightness{}
}

type brightItem struct {
	input string
	b     *brightness
}

func (i *brightItem) Icon() fyne.Resource {
	return wmtheme.BrightnessIcon
}

func (i *brightItem) Title() string {
	if startsWith(i.input, "down") {
		return "Brightness down"
	} else if val, err := strconv.Atoi(i.input); err == nil {
		return fmt.Sprintf("Brightness %d%%", val)
	}

	return "Brightness up"
}

func (i *brightItem) Launch() {
	if startsWith(i.input, "down") {
		i.b.offsetValue(-5)
	} else if val, err := strconv.Atoi(i.input); err == nil {
		i.b.setValue(val)
	}

	i.b.offsetValue(5)
}
