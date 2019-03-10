// +build !linux ci

package desktop

import (
	"log"

	"fyne.io/fyne"
	_ "fyne.io/fyne/test"
)

func isEmbedded() bool {
	return true
}

// newDesktopWindow creates a test window in memory for automated testing.
func newDesktopWindow(a fyne.App) fyne.Window {
	if !isEmbedded() {
		log.Fatalln("Fyne Desktop currenly only works on Linux")
	}

	return createWindow(a)
}
