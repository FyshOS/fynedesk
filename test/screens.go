package test

import "fyshos.com/fynedesk"

type testScreensProvider struct {
	screens []*fynedesk.Screen
	primary *fynedesk.Screen
	active  *fynedesk.Screen
}

// NewScreensProvider returns a simple screen manager for the specified screens
func NewScreensProvider(screens ...*fynedesk.Screen) fynedesk.ScreenList {
	if screens == nil {
		screens = []*fynedesk.Screen{{Name: "Screen0", X: 0, Y: 0, Width: 2000, Height: 1000, Scale: 1.0}}
	}
	return &testScreensProvider{screens: screens, active: screens[0], primary: screens[0]}
}

func (tsp *testScreensProvider) RefreshScreens() {
}

func (tsp *testScreensProvider) AddChangeListener(func()) {
	// no-op
}

func (tsp *testScreensProvider) Screens() []*fynedesk.Screen {
	return tsp.screens
}

func (tsp *testScreensProvider) SetActive(s *fynedesk.Screen) {
	tsp.active = s
}

func (tsp *testScreensProvider) Active() *fynedesk.Screen {
	return tsp.active
}

func (tsp *testScreensProvider) Primary() *fynedesk.Screen {
	return tsp.primary
}

func (tsp *testScreensProvider) Scale() float32 {
	return 1.0
}

func (tsp *testScreensProvider) ScreenForWindow(win fynedesk.Window) *fynedesk.Screen {
	return tsp.Screens()[0]
}

func (tsp *testScreensProvider) ScreenForGeometry(x int, y int, width int, height int) *fynedesk.Screen {
	return tsp.Screens()[0]
}
