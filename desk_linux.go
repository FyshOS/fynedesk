// +build !ci

package desktop

import "os"

import "fyne.io/fyne"

var desk fyne.Window

func isEmbedded() bool {
	env := os.Getenv("DISPLAY")
	if env != "" {
		return true
	}

	env = os.Getenv("WAYLAND_DISPLAY")
	return env != ""
}
