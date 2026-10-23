package status

import (
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyshos.com/tyde/internal/modhelper"

	"github.com/mafik/pulseaudio"

	"fyshos.com/tyde"
	wmtheme "fyshos.com/tyde/theme"
)

var soundMeta = tyde.ModuleMetadata{
	Name:        "Sound",
	NewInstance: newSound,
}

type sound struct {
	bar    *widget.ProgressBar
	client *pulseaudio.Client
	mute   *scrollButton
}

func newSound() tyde.Module {
	return &sound{}
}

func (b *sound) LaunchSuggestions(input string) []tyde.LaunchSuggestion {
	if _, err := b.value(); err != nil {
		return nil // don't load if not present
	}

	lower := strings.ToLower(input)
	val := lower
	switch {
	case startsWith(lower, "volume "):
		val = "" // still being typed: "vol", "volum" - nothing to act on yet
		if len(lower) > len("volume ") {
			val = lower[len("volume "):]
		}
	case startsWith(lower, "vol "):
		val = ""
		if len(lower) > len("vol ") {
			val = lower[len("vol "):]
		}
	case lower == "mute" || lower == "unmute":
		// the word is the value
	default:
		return nil
	}

	item := b.item(val)
	if item == nil {
		return nil
	}
	return []tyde.LaunchSuggestion{item}
}

func (b *sound) item(val string) tyde.LaunchSuggestion {
	val = strings.ToLower(strings.TrimSpace(val))

	if _, err := strconv.Atoi(val); err != nil {
		if !startsWith(val, "u") && !startsWith(val, "d") && val != "mute" && val != "unmute" {
			return nil
		}
	}

	return &volItem{input: val, s: b}
}

func (b *sound) Shortcuts() map[*tyde.Shortcut]func() {
	return map[*tyde.Shortcut]func(){
		tyde.NewShortcut("Mute Sound", tyde.KeyVolumeMute, tyde.AnyModifier): func() {
			b.toggleMute()
		},
		tyde.NewShortcut("Increase Sound Volume", tyde.KeyVolumeDown, tyde.AnyModifier): func() {
			b.offsetValue(-5)
		},
		tyde.NewShortcut("Reduce Sound Volume", tyde.KeyVolumeUp, tyde.AnyModifier): func() {
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
	b.mute = newScrollButton(wmtheme.SoundHighIcon)
	b.mute.scroll = func(f float32) {
		if b.muted() {
			return
		}

		b.offsetValue(int(f / 10))
	}
	b.mute.Importance = widget.LowImportance
	b.mute.OnTapped = b.toggleMute
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
	return container.New(&handleNarrow{}, b.mute, sound)
}

// Metadata returns ModuleMetadata
func (b *sound) Metadata() tyde.ModuleMetadata {
	return soundMeta
}

// Commands describes the volume commands, offered only where there is a mixer
// to talk to - the same test LaunchSuggestions makes before matching.
func (b *sound) Commands() []tyde.Command {
	if _, err := b.value(); err != nil {
		return nil
	}

	c1 := tyde.Command{
		Name:        "volume",
		Description: "Set or step the output volume",
		Args:        "a percentage from 0 to 100, or 'up' / 'down' to step by 5%",
		Examples:    []string{"volume 30", "volume down"},
	}
	modhelper.AnswerWith(&c1, b.item)

	c2 := tyde.Command{
		Name:        "mute",
		Description: "Mute the sound output",
	}
	modhelper.AnswerWith(&c2, b.always("mute"))

	c3 := tyde.Command{
		Name:        "unmute",
		Description: "Unmute the sound output",
	}
	modhelper.AnswerWith(&c3, b.always("unmute"))
	return []tyde.Command{c1, c2, c3}
}

// always ignores what followed the command name: "mute" means the one thing
// whether or not anything was typed after it.
func (b *sound) always(val string) func(string) tyde.LaunchSuggestion {
	return func(string) tyde.LaunchSuggestion {
		return b.item(val)
	}
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

func (b *sound) updateIcon(vol int, mute bool) {
	if b.mute == nil {
		return // a command can arrive before, or without, a status area to update
	}

	if mute {
		b.mute.SetIcon(wmtheme.MuteIcon)
	} else {
		if vol <= 20 {
			b.mute.SetIcon(wmtheme.SoundLowIcon)
		} else if vol <= 60 {
			b.mute.SetIcon(wmtheme.SoundMidIcon)
		} else {
			b.mute.SetIcon(wmtheme.SoundHighIcon)
		}
	}
}

type volItem struct {
	input string
	s     *sound
}

func (i *volItem) Icon() fyne.Resource {
	return wmtheme.SoundHighIcon
}

func (i *volItem) Title() string {
	if _, err := strconv.Atoi(i.input); err == nil {
		return "Volume " + i.input + "%"
	} else if i.input == "mute" {
		return "Mute volume"
	} else if i.input == "unmute" {
		return "Unmute volume"
	} else if startsWith(i.input, "u") {
		return "Volume up"
	} else if startsWith(i.input, "d") {
		return "Volume down"
	}

	return ""
}

func (i *volItem) Launch() {
	if i.input == "mute" {
		i.s.setMute(true)
	} else if i.input == "unmute" {
		i.s.setMute(false)
	} else if startsWith(i.input, "u") {
		i.s.offsetValue(5)
	} else if startsWith(i.input, "d") {
		i.s.offsetValue(-5)
	} else if val, err := strconv.Atoi(i.input); err == nil {
		if val < 0 {
			val = 0
		} else if val > 100 {
			val = 100
		}
		i.s.setValue(val)
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
	return strings.Index(haystack, needle) == 0
}
