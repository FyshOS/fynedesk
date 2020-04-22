package ui

import (
	"fyne.io/fynedesk"
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

func (e *embededWM) Close() {
	// no-op, just allow exit
}
