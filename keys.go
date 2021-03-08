package fynedesk

import (
	"fyne.io/fyne/v2"

	deskDriver "fyne.io/fyne/v2/driver/desktop"
)

const (
	// AnyModifier is the shortcut modifier to use if the shortcut should always trigger - use sparingly
	AnyModifier deskDriver.Modifier = 0
	// UserModifier is the shortcut modifier to use if the shortcut should respect user preference.
	// This will be offered as a choice of Alt or Super (Command)
	UserModifier deskDriver.Modifier = deskDriver.SuperModifier << 1

	// KeyBrightnessDown is the virtual keyboard key for reducing brightness
	KeyBrightnessDown fyne.KeyName = "BrightnessDown"
	// KeyBrightnessUp is the virtual keyboard key for increasing brightness
	KeyBrightnessUp fyne.KeyName = "BrightnessUp"

	// KeyVolumeMute is the virtual keyboard key for muting sound
	KeyVolumeMute fyne.KeyName = "VolumeMute"
	// KeyVolumeDown is the virtual keyboard key for reducing sound level
	KeyVolumeDown fyne.KeyName = "VolumeDown"
	// KeyVolumeUp is the virtual keyboard key for increasing sound level
	KeyVolumeUp fyne.KeyName = "VolumeUp"
)

// Declare conformity with Shortcut interface
var _ fyne.Shortcut = (*Shortcut)(nil)

// Shortcut defines a keyboard shortcut that can be configured by the user
type Shortcut struct {
	fyne.KeyName
	deskDriver.Modifier
	Name string
}

// ShortcutName gets the name of this shortcut - this should be user presentable
func (s *Shortcut) ShortcutName() string {
	return s.Name
}

// NewShortcut creates a keyboard shortcut that can be configured by the user
func NewShortcut(name string, key fyne.KeyName, mods deskDriver.Modifier) *Shortcut {
	s := &Shortcut{Name: name}
	s.KeyName = key
	s.Modifier = mods
	return s
}
