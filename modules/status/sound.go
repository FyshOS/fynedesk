package status

import (
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/mafik/pulseaudio"

	"fyne.io/fynedesk"
	wmtheme "fyne.io/fynedesk/theme"
)

var soundMeta = fynedesk.ModuleMetadata{
	Name:        "Sound",
	NewInstance: newSound,
}

type sound struct {
	bar    *widget.ProgressBar
	client *pulseaudio.Client
	mute   *widget.Button
}

func newSound() fynedesk.Module {
	return &sound{}
}

func (b *sound) LaunchSuggestions(input string) []fynedesk.LaunchSuggestion {
	if _, err := b.value(); err != nil {
		return nil // don't load if not present
	}

	lower := strings.ToLower(input)
	matches := false
	val := lower
	if startsWith(lower, "volume ") {
		matches = true
		if len(lower) > 7 {
			val = lower[7:]
		} else {
			val = ""
		}
	} else if startsWith(lower, "vol ") {
		matches = true
		if len(lower) > 4 {
			val = lower[4:]
		} else {
			val = ""
		}
	} else if startsWith(lower, "mute") || startsWith(lower, "unmute") {
		matches = true
	}

	if matches {
		return []fynedesk.LaunchSuggestion{&volItem{input: val, s: b}}
	}

	return nil
}

func (b *sound) Shortcuts() map[*fynedesk.Shortcut]func() {
	return map[*fynedesk.Shortcut]func(){
		fynedesk.NewShortcut("Mute Sound", fynedesk.KeyVolumeMute, fynedesk.AnyModifier): func() {
			b.toggleMute()
		},
		fynedesk.NewShortcut("Increase Sound Volume", fynedesk.KeyVolumeDown, fynedesk.AnyModifier): func() {
			b.offsetValue(-5)
		},
		fynedesk.NewShortcut("Reduce Sound Volume", fynedesk.KeyVolumeUp, fynedesk.AnyModifier): func() {
			b.offsetValue(5)
		},
	}
}

// StatusAreaWidget builds the widget
func (b *sound) StatusAreaWidget() fyne.CanvasObject {
	if err := b.setup(); err != nil {
		fyne.LogError("Unable to start sound module", err)
		return nil
	}

	b.bar = &widget.ProgressBar{Max: 100}
	b.mute = &widget.Button{Icon: wmtheme.SoundIcon, Importance: widget.LowImportance, OnTapped: b.toggleMute}
	if b.muted() {
		b.mute.SetIcon(wmtheme.MuteIcon)
	}

	less := &widget.Button{Icon: theme.ContentRemoveIcon(), Importance: widget.LowImportance, OnTapped: func() {
		b.offsetValue(-5)
	}}

	more := &widget.Button{Icon: theme.ContentAddIcon(), Importance: widget.LowImportance, OnTapped: func() {
		b.offsetValue(5)
	}}

	sound := container.NewBorder(nil, nil, less, more, b.bar)

	go b.offsetValue(0)
	return container.NewBorder(nil, nil, b.mute, nil, sound)
}

// Metadata returns ModuleMetadata
func (b *sound) Metadata() fynedesk.ModuleMetadata {
	return soundMeta
}

func (b *sound) offsetValue(diff int) {
	currVal, err := b.value()
	if err != nil {
		fyne.LogError("Failed to get volume", err)
		return
	}
	value := currVal + diff

	if value < 0 {
		value = 0
	} else if value > 100 {
		value = 100
	}

	b.setValue(value)
}

type volItem struct {
	input string
	s     *sound
}

func (i *volItem) Icon() fyne.Resource {
	return wmtheme.SoundIcon
}

func (i *volItem) Title() string {
	if startsWith(i.input, "up") {
		return "Volume up"
	} else if startsWith(i.input, "down") {
		return "Volume down"
	} else if val, err := strconv.Atoi(i.input); err == nil {
		return "Volume " + strconv.Itoa(val) + "%"
	}

	if i.s.muted() {
		return "Unmute volume"
	}
	return "Mute volume"
}

func (i *volItem) Launch() {
	if startsWith(i.input, "up") {
		i.s.offsetValue(5)
	} else if startsWith(i.input, "down") {
		i.s.offsetValue(-5)
	} else if val, err := strconv.Atoi(i.input); err == nil {
		if val < 0 {
			val = 0
		} else if val > 100 {
			val = 100
		}
		i.s.setValue(val)
	} else {
		i.s.toggleMute()
	}
}

func startsWith(haystack, needle string) bool {
	if haystack == "" {
		return false
	}
	if haystack == needle {
		return true
	}
	if len(haystack) < len(needle) {
		return haystack == needle[:len(haystack)]
	}
	return strings.IndexAny(haystack, needle) == 0
}
