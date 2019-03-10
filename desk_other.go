// +build !linux ci

package desktop

import "log"

import "fyne.io/fyne"
import _ "fyne.io/fyne/test"

func isEmbedded() bool {
	return true
}

// newDesktopWindow creates a test window in memory for automated testing.
func newDesktopWindow(app fyne.App) fyne.Window {
	log.Fatalln("Fyne Desktop currenly only works on Linux")

	return nil
}
