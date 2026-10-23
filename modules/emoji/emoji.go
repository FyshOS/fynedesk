// Package emoji adds an emoji picker to Tyde: a grouped grid of emoji, where
// picking one copies it to the clipboard ready to paste. It is reachable four
// ways - the Super+. keyboard shortcut, an "emoji" entry in the app launcher, an
// icon in the widget panel, and the Emoji Picker application entry (which runs
// "tyde_ctl emoji") - because which one feels natural depends entirely on
// whether your hands are on the keyboard at the time.
//
// Typing the emoji straight into the window you came from, the way the Windows
// and macOS pickers do, is drafted in insert.go but not currently wired in.
package emoji

import (
	_ "embed"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyshos.com/tyde/internal/modhelper"

	"fyshos.com/tyde"
)

// ModuleName is the registered name of the emoji picker module.
const ModuleName = "Emoji Picker"

// launchAliases are the words that summon the picker from the app launcher.
var launchAliases = []string{"emoji", "emoticon", "smiley", "symbol"}

//go:embed emoji.svg
var emojiSvg []byte

// Icon is the picker's themed icon, exported so the launcher suggestion, the
// widget panel and the desktop entry all show the same glyph.
var Icon fyne.Resource = theme.NewThemedResource(&fyne.StaticResource{
	StaticName:    "emoji.svg",
	StaticContent: emojiSvg,
})

var meta = tyde.ModuleMetadata{
	Name:        ModuleName,
	NewInstance: newEmoji,
}

// module is the module instance. It owns the single picker overlay.
type module struct {
	picker *picker
}

func newEmoji() tyde.Module {
	m := &module{}
	m.picker = &picker{}
	return m
}

func (m *module) Metadata() tyde.ModuleMetadata {
	return meta
}

func (m *module) Destroy() {
	if m.picker != nil {
		m.picker.hide()
	}
}

// Shortcuts binds the picker to Super+. - the same chord Windows uses for its
// emoji panel, and unclaimed elsewhere in Tyde.
func (m *module) Shortcuts() map[*tyde.Shortcut]func() {
	return map[*tyde.Shortcut]func(){
		tyde.NewShortcut("Show Emoji Picker", fyne.KeyPeriod, tyde.UserModifier): func() {
			m.picker.toggle()
		},
	}
}

// StatusAreaWidget puts a picker button in the widget panel for the times the
// mouse is already in hand.
func (m *module) StatusAreaWidget() fyne.CanvasObject {
	return &widget.Button{
		Icon: Icon, Importance: widget.LowImportance,
		OnTapped: func() { m.picker.toggle() },
	}
}

// LaunchSuggestions offers the picker when the launcher input starts to look
// like one of its aliases, and searches emoji directly once there is something
// to search for - so "emoji fire" puts 🔥 one Enter away.
func (m *module) LaunchSuggestions(input string) []tyde.LaunchSuggestion {
	lower := strings.ToLower(strings.TrimSpace(input))
	if lower == "" {
		return nil
	}

	for _, alias := range launchAliases {
		query, isSearch := strings.CutPrefix(lower, alias+" ")
		switch {
		case isSearch:
			return m.searchSuggestions(query)
		case strings.HasPrefix(alias, lower):
			return []tyde.LaunchSuggestion{&openItem{m: m}}
		}
	}
	return nil
}

// searchSuggestions offers the best matches for a typed name, each of which
// delivers that emoji without opening the picker at all.
func (m *module) searchSuggestions(query string) []tyde.LaunchSuggestion {
	const maxSuggestions = 5

	results := Search(query)
	if len(results) == 0 {
		return []tyde.LaunchSuggestion{&openItem{m: m}}
	}
	if len(results) > maxSuggestions {
		results = results[:maxSuggestions]
	}

	out := make([]tyde.LaunchSuggestion, 0, len(results))
	for _, e := range results {
		out = append(out, &emojiItem{m: m, emoji: e})
	}
	return out
}

// ShowPicker opens the picker, as the Emoji Picker application entry and
// "tyde_ctl emoji" do. Must be called on the UI thread.
func (m *module) ShowPicker() {
	m.picker.show()
}

// Commands offers the picker, and the emoji search behind it, to tyde_ctl and
// the assistant.
func (m *module) Commands() []tyde.Command {
	c := tyde.Command{
		Name:        launchAliases[0],
		Description: "Open the emoji picker, or copy an emoji to the clipboard by name",
		Args:        "optional: the name of the emoji to copy",
		Examples:    []string{"emoji", "emoji fire"},
	}

	modhelper.AnswerWith(&c, m.item)
	return []tyde.Command{c}
}

func (m *module) item(query string) tyde.LaunchSuggestion {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return &openItem{m: m}
	}

	return m.searchSuggestions(query)[0]
}

// openItem is the launcher suggestion that opens the picker.
type openItem struct {
	m *module
}

func (i *openItem) Icon() fyne.Resource { return Icon }

func (i *openItem) Title() string { return "Emoji Picker" }

func (i *openItem) Launch() { i.m.picker.show() }

// emojiItem is a launcher suggestion for one named emoji, delivered straight
// from the launcher.
type emojiItem struct {
	m     *module
	emoji Emoji
}

func (i *emojiItem) Icon() fyne.Resource { return Icon }

func (i *emojiItem) Title() string { return i.emoji.Character + "  " + i.emoji.Name }

// Launch copies the emoji to the clipboard, ready to paste.
func (i *emojiItem) Launch() {
	fyne.CurrentApp().Clipboard().SetContent(i.emoji.Character)
}
