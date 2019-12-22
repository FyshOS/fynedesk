package notify

import "fyne.io/fyne"

// MouseNotify is an interface that lets separate packages like wm tell the desktop that the cursor has moved to or from the desktop canvas.
type MouseNotify interface {
	MouseInNotify(fyne.Position)
	MouseOutNotify()
}
