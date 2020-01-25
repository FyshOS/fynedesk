package ui

import "fyne.io/desktop"

type embeddedScreensProvider struct {
	screens []*desktop.Screen
}

func (esp embeddedScreensProvider) RefreshScreens() {
	return
}

func (esp embeddedScreensProvider) AddChangeListener(func()) {
	// no-op
}

func (esp embeddedScreensProvider) Screens() []*desktop.Screen {
	return esp.screens
}

func (esp embeddedScreensProvider) Active() *desktop.Screen {
	return esp.Screens()[0]
}

func (esp embeddedScreensProvider) Primary() *desktop.Screen {
	return esp.Screens()[0]
}

func (esp embeddedScreensProvider) ScreenForWindow(win desktop.Window) *desktop.Screen {
	return esp.Screens()[0]
}

func (esp embeddedScreensProvider) ScreenForGeometry(x int, y int, width int, height int) *desktop.Screen {
	return esp.Screens()[0]
}

// NewEmbeddedScreensProvider returns a screen provider for use in embedded desktop mode
func newEmbeddedScreensProvider() desktop.ScreenList {
	return &embeddedScreensProvider{[]*desktop.Screen{{Name: "(Embedded)", X: 0, Y: 0,
		Width: 1280, Height: 1024, Scale: 1.0}}}
}
