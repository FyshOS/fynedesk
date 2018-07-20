// +build !linux ci

package driver

import "github.com/fyne-io/fyne"
import "github.com/fyne-io/fyne/test"

// NewApp creates a new headless app to test the desktop code
func NewApp() fyne.App {
	return test.NewApp()
}
