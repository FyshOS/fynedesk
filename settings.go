package desktop

import "os"

// DeskSettings describes the configuration options available for Fyne desktop
type DeskSettings interface {
	IconTheme() string
}

type deskSettings struct {
	iconTheme string
}

func (d *deskSettings) IconTheme() string {
	return d.iconTheme
}

func (d *deskSettings) load() {
	env := os.Getenv("FYNEDESK_ICONTHEME")
	if env == "" {
		d.iconTheme = "hicolor"
	}

	d.iconTheme = env
}

// NewDeskSettings loads the user's preferences from environment or config
func NewDeskSettings() DeskSettings {
	settings := &deskSettings{}
	settings.load()

	return settings
}
