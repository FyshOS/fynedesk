package ui

import (
	"image"

	"fyne.io/fyne/v2"

	"fyshos.com/fynedesk"
)

type embededWM struct {
	windows []fynedesk.Window
}

func (e *embededWM) AddWindow(win fynedesk.Window) {
	e.windows = append(e.windows, win)
}

func (e *embededWM) RaiseToTop(fynedesk.Window) {
	// no-op
}

func (e *embededWM) RemoveWindow(win fynedesk.Window) {
	for i, w := range e.windows {
		if w != win {
			continue
		}

		e.windows = append(e.windows[:i], e.windows[i+1:]...)
		return
	}
}

func (e *embededWM) Run() {
}

func (e *embededWM) ShowOverlay(w fyne.Window, s fyne.Size, p fyne.Position) {
	w.Resize(s)
	w.Show()
}

func (e *embededWM) ShowMenuOverlay(*fyne.Menu, fyne.Size, fyne.Position) {
	// no-op, handled by desktop in embed mode
}

func (e *embededWM) TopWindow() fynedesk.Window {
	if len(e.windows) == 0 {
		return nil
	}

	return e.windows[len(e.windows)-1]
}

func (e *embededWM) Windows() []fynedesk.Window {
	return e.windows
}

func (e *embededWM) AddStackListener(fynedesk.StackListener) {
	// no stack
}

func (e *embededWM) Blank() {
	// no-op, we don't control screen brightness
}

func (e *embededWM) Capture() image.Image {
	return nil // would mean accessing the underling OS screen functions...
}

func (e *embededWM) Close() {
	windows := fyne.CurrentApp().Driver().AllWindows()
	if len(windows) > 0 {
		windows[0].Close() // ensure our root is asked to close as well
	}
}
