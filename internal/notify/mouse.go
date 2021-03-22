package notify

import "fyne.io/fyne/v2"

// MouseNotify is an interface that can be used by objects interested in when the mouse enters or exits the desktop
type MouseNotify interface {
	MouseInNotify(fyne.Position)
	MouseOutNotify()
}
