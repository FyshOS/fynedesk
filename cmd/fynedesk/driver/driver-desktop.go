// +build linux,!ci

package driver

import "github.com/fyne-io/fyne"
import "github.com/fyne-io/fyne/desktop"

// NewApp creates a new desktop app to run the desktop
func NewApp() fyne.App {
	return desktop.NewApp()
}
