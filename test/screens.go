package test

import "fyne.io/fynedesk"

type testScreensProvider struct {
	screens []*fynedesk.Screen
	primary *fynedesk.Screen
	active  *fynedesk.Screen
}

func NewScreensProvider(screens ...*fynedesk.Screen) *testScreensProvider {
	return &testScreensProvider{screens: screens}
}

func (tsp testScreensProvider) RefreshScreens() {
	return
}

func (tsp testScreensProvider) AddChangeListener(func()) {
	// no-op
}

func (tsp testScreensProvider) Screens() []*fynedesk.Screen {
	return tsp.screens
}

func (tsp testScreensProvider) Active() *fynedesk.Screen {
	return tsp.screens[0]
}

func (tsp testScreensProvider) Primary() *fynedesk.Screen {
	return tsp.screens[0]
}

func (tsp testScreensProvider) Scale() float32 {
	return 1.0
}

func (tsp testScreensProvider) ScreenForWindow(win fynedesk.Window) *fynedesk.Screen {
	return tsp.Screens()[0]
}

func (tsp testScreensProvider) ScreenForGeometry(x int, y int, width int, height int) *fynedesk.Screen {
	return tsp.Screens()[0]
}
