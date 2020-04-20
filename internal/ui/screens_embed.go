package ui

import "fyne.io/fynedesk"

type embeddedScreensProvider struct {
	screens []*fynedesk.Screen
}

func (esp embeddedScreensProvider) RefreshScreens() {
	return
}

func (esp embeddedScreensProvider) AddChangeListener(func()) {
	// no-op
}

func (esp embeddedScreensProvider) Screens() []*fynedesk.Screen {
	return esp.screens
}

func (esp embeddedScreensProvider) Active() *fynedesk.Screen {
	return esp.Screens()[0]
}

func (esp embeddedScreensProvider) Primary() *fynedesk.Screen {
	return esp.Screens()[0]
}

func (esp embeddedScreensProvider) ScreenForWindow(win fynedesk.Window) *fynedesk.Screen {
	return esp.Screens()[0]
}

func (esp embeddedScreensProvider) ScreenForGeometry(x int, y int, width int, height int) *fynedesk.Screen {
	return esp.Screens()[0]
}

// NewEmbeddedScreensProvider returns a screen provider for use in embedded desktop mode
func newEmbeddedScreensProvider() fynedesk.ScreenList {
	return &embeddedScreensProvider{[]*fynedesk.Screen{{Name: "(Embedded)", X: 0, Y: 0,
		Width: 1280, Height: 1024, Scale: 1.0}}}
}
