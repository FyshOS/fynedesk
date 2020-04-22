package test

import (
	"fyne.io/fyne"
	"fyne.io/fyne/test"

	"fyne.io/fynedesk"
)

type Desktop struct {
	settings fynedesk.DeskSettings
	icons    fynedesk.ApplicationProvider
	screens  fynedesk.ScreenList
	wm       fynedesk.WindowManager
}

func NewDesktop() *Desktop {
	screens := &testScreensProvider{screens: []*fynedesk.Screen{{Name: "Screen0", X: 0, Y: 0, Width: 2000, Height: 1000, Scale: 1.0}}}
	return &Desktop{settings: &Settings{}, icons: &testAppProvider{}, screens: screens}
}

func NewDesktopWithWM(wm fynedesk.WindowManager) *Desktop {
	desk := NewDesktop()
	desk.wm = wm
	return desk
}

func (*Desktop) ContentSizePixels(screen *fynedesk.Screen) (uint32, uint32) {
	return uint32(320), uint32(240)
}

func (td *Desktop) IconProvider() fynedesk.ApplicationProvider {
	return td.icons
}

func (td *Desktop) SetIconProvider(icons fynedesk.ApplicationProvider) {
	td.icons = icons
}

func (*Desktop) Modules() []fynedesk.Module {
	return nil
}

func (*Desktop) Root() fyne.Window {
	return test.NewWindow(nil)
}

func (*Desktop) Run() {
}

func (*Desktop) RunApp(app fynedesk.AppData) error {
	return app.Run([]string{}) // no added env
}

func (td *Desktop) Screens() fynedesk.ScreenList {
	return td.screens
}

func (td *Desktop) Settings() fynedesk.DeskSettings {
	return td.settings
}

func (td *Desktop) WindowManager() fynedesk.WindowManager {
	return td.wm
}
