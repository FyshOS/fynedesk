// +build !linux ci

package desktop

import "log"
import "runtime"

import "fyne.io/fyne"
import _ "fyne.io/fyne/test"

func isEmbedded() bool {
	return true
}

// newDesktopWindow creates a test window in memory for automated testing.
func newDesktopWindow(app fyne.App) fyne.Window {
	if runtime.GOOS != "linux" {
		log.Println("Fyne Desktop currenly only works on Linux")
	}
	return app.NewWindow("Fyne Desktop")
}

func initInput() {
}
