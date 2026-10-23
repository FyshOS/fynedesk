package status

import (
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyshos.com/tyde/internal/modhelper"
	"github.com/FyshOS/dryvers"

	"fyshos.com/tyde"
	wmtheme "fyshos.com/tyde/theme"
)

// BrightnessModuleName is the name of the screen brightness module.
const BrightnessModuleName = "Brightness"

var brightnessMeta = tyde.ModuleMetadata{
	Name:        BrightnessModuleName,
	NewInstance: newBrightness,
}

// Brightness is a progress bar module to modify screen brightness
type brightness struct {
	bright *dryvers.Brightness

	bar *widget.ProgressBar
}

func (b *brightness) Destroy() {
}

func (b *brightness) offsetValue(diff int) {
	floatVal, _ := b.bright.Get()
	if floatVal <= 0.01 { // don't start doing 6, 11 etc just because we were on 1 (min)
		floatVal = 0
	}
	value := int(floatVal*100) + diff

	b.setValue(value)
}

func (b *brightness) setValue(value int) {
	if value < 1 {
		value = 1
	} else if value > 100 {
		value = 100
	}

	_ = b.bright.Set(float64(value) / 100)

	newVal, _ := b.bright.Get()
	if b.bar != nil { // a command can arrive before, or without, a status area to update
		b.bar.SetValue(newVal)
	}
}

func (b *brightness) LaunchSuggestions(input string) []tyde.LaunchSuggestion {
	if val, err := b.bright.Get(); err != nil || val == 0 {
		return nil // don't load if not present
	}

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

	if !matches {
		return nil
	}

	item := b.item(val)
	if item == nil {
		return nil
	}
	return []tyde.LaunchSuggestion{item}
}

func (b *brightness) item(val string) tyde.LaunchSuggestion {
	val = strings.ToLower(strings.TrimSpace(val))

	if _, err := strconv.Atoi(val); err != nil {
		if !startsWith(val, "d") && !startsWith(val, "u") {
			return nil
		}
	}

	return &brightItem{input: val, b: b}
}

func (b *brightness) Metadata() tyde.ModuleMetadata {
	return brightnessMeta
}

// Commands describes the brightness command, offered only where there is a backlight.
func (b *brightness) Commands() []tyde.Command {
	if val, err := b.bright.Get(); err != nil || val == 0 {
		return nil
	}

	c := tyde.Command{
		Name:        "brightness",
		Description: "Set or step the screen backlight brightness",
		Args:        "a percentage from 1 to 100, or 'up' / 'down' to step by 5%",
		Examples:    []string{"brightness 50", "brightness up"},
	}

	modhelper.AnswerWith(&c, b.item)
	return []tyde.Command{c}
}

func (b *brightness) Shortcuts() map[*tyde.Shortcut]func() {
	return map[*tyde.Shortcut]func(){
		tyde.NewShortcut("Increase Screen Brightness", tyde.KeyBrightnessDown, tyde.AnyModifier): func() {
			b.offsetValue(-5)
		},
		tyde.NewShortcut("Reduce Screen Brightness", tyde.KeyBrightnessUp, tyde.AnyModifier): func() {
			b.offsetValue(5)
		},
	}
}

func (b *brightness) StatusAreaWidget() fyne.CanvasObject {
	if val, err := b.bright.Get(); err != nil || val == 0 {
		return nil
	}

	b.bar = widget.NewProgressBar()
	brightnessIcon := newScrollIcon(wmtheme.BrightnessIcon)
	brightnessIcon.scroll = func(f float32) {
		b.offsetValue(int(f / 10))
	}

	prop := canvas.NewRectangle(color.Transparent)
	prop.SetMinSize(brightnessIcon.MinSize().Add(fyne.NewSize(theme.Padding()*4, 0)))
	icon := container.NewCenter(prop, brightnessIcon)

	less := &widget.Button{Icon: theme.ContentRemoveIcon(), Importance: widget.LowImportance, OnTapped: func() {
		b.offsetValue(-5)
	}}

	more := &widget.Button{Icon: theme.ContentAddIcon(), Importance: widget.LowImportance, OnTapped: func() {
		b.offsetValue(5)
	}}

	bright := container.NewBorder(nil, nil, less, more, b.bar)

	go b.offsetValue(0)
	return container.New(&handleNarrow{}, icon, bright)
}

// newBrightness creates a new module that will show screen brightness in the status area
func newBrightness() tyde.Module {
	return &brightness{bright: dryvers.NewBrightness()}
}

type brightItem struct {
	input string
	b     *brightness
}

func (i *brightItem) Icon() fyne.Resource {
	return wmtheme.BrightnessIcon
}

func (i *brightItem) Title() string {
	if _, err := strconv.Atoi(i.input); err == nil {
		return "Brightness " + i.input + "%"
	} else if startsWith(i.input, "d") {
		return "Brightness down"
	} else if startsWith(i.input, "u") {
		return "Brightness up"
	}

	return ""
}

func (i *brightItem) Launch() {
	if val, err := strconv.Atoi(i.input); err == nil {
		i.b.setValue(val)
	} else if startsWith(i.input, "d") {
		i.b.offsetValue(-5)
	} else if startsWith(i.input, "u") {
		i.b.offsetValue(5)
	}
}
