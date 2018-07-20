// +build !linux ci

package desktop

import "github.com/fyne-io/fyne"
import _ "github.com/fyne-io/fyne/test"

func isEmbedded() bool {
	return true
}

// NewDesktopWindow creates a test window in memory for automated testing.
func NewDesktopWindow(app fyne.App) fyne.Window {
	return app.NewWindow("Fyne Desktop")
}

func initInput() {
}
