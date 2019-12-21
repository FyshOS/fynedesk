package notify

import "fyne.io/fyne"

type MouseNotify interface {
	MouseInNotify(fyne.Position)
	MouseOutNotify()
}
