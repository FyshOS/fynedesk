package desktop

import (
	"os"

	"fyne.io/fyne"
)

// DeskSettings describes the configuration options available for Fyne desktop
type DeskSettings interface {
	IconTheme() string
	Background() string
}

type deskSettings struct {
	iconTheme  string
	background string
}

func (d *deskSettings) IconTheme() string {
	return d.iconTheme
}

func (d *deskSettings) Background() string {
	return d.background
}

func (d *deskSettings) load() {
	env := os.Getenv("FYNEDESK_ICONTHEME")
	if env == "" {
		d.iconTheme = "hicolor"
	} else {
		d.iconTheme = env
	}

	env = os.Getenv("FYNEDESK_BACKGROUND")
	if env != "" {
		d.background = env
	} else {
		d.background = fyne.CurrentApp().Preferences().String("background")
	}
}

// NewDeskSettings loads the user's preferences from environment or config
func NewDeskSettings() DeskSettings {
	settings := &deskSettings{}
	settings.load()

	return settings
}
