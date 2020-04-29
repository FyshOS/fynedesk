package ui

import "fyne.io/fynedesk"

type embeddedScreensProvider struct {
	active  *fynedesk.Screen
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

func (esp embeddedScreensProvider) SetActive(s *fynedesk.Screen) {
	esp.active = s
}

func (esp embeddedScreensProvider) Active() *fynedesk.Screen {
	return esp.active
}

func (esp embeddedScreensProvider) Primary() *fynedesk.Screen {
	return esp.Screens()[0]
}

func (esp embeddedScreensProvider) ScreenForWindow(_ fynedesk.Window) *fynedesk.Screen {
	return esp.Screens()[0]
}

func (esp embeddedScreensProvider) ScreenForGeometry(_ fynedesk.Geometry) *fynedesk.Screen {
	return esp.Screens()[0]
}

// NewEmbeddedScreensProvider returns a screen provider for use in embedded desktop mode
func newEmbeddedScreensProvider() fynedesk.ScreenList {
	screen := &fynedesk.Screen{Name: "(Embedded)", Scale: 1.0,
		Geometry: fynedesk.Geometry{X: 0, Y: 0, Width: 1280, Height: 1024}}
	return &embeddedScreensProvider{active: screen, screens: []*fynedesk.Screen{screen}}
}
