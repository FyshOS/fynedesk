package desktop

import "fyne.io/fyne"

// WindowManager describes a full window manager which may be loaded as part of the setup.
type WindowManager interface {
	Close()
	SetRoot(window fyne.Window)
}