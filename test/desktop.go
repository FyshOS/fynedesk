package test

import (
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"fyne.io/fynedesk"
)

// Desktop is an in-memory implementation for test purposes
type Desktop struct {
	settings fynedesk.DeskSettings
	icons    fynedesk.ApplicationProvider
	screens  fynedesk.ScreenList
	wm       fynedesk.WindowManager
}

// NewDesktop returns a new in-memory desktop instance
func NewDesktop() *Desktop {
	screen := &fynedesk.Screen{Name: "Screen0", X: 0, Y: 0, Width: 2000, Height: 1000, Scale: 1.0}
	screens := &testScreensProvider{screens: []*fynedesk.Screen{screen}, active: screen, primary: screen}
	return &Desktop{settings: &Settings{}, icons: &testAppProvider{}, screens: screens}
}

// NewDesktopWithWM returns a new in-memory desktop instance using the specified window manager
func NewDesktopWithWM(wm fynedesk.WindowManager) *Desktop {
	desk := NewDesktop()
	desk.wm = wm
	return desk
}

// AddShortcut is called from modules that wish to register keyboard handlers
func (*Desktop) AddShortcut(shortcut *fynedesk.Shortcut, handler func()) {
	// TODO
}

// Capture the desktop to an image. Our test code cowardly refuses to do this.
func (*Desktop) Capture() image.Image {
	return nil // could be implemented if required for testing
}

// ContentSizePixels returns a default value for how much space maximised apps should use
func (*Desktop) ContentSizePixels(_ *fynedesk.Screen) (uint32, uint32) {
	return uint32(320), uint32(240)
}

// IconProvider returns the icon provider, by default it uses a simple in-memory implementation
func (td *Desktop) IconProvider() fynedesk.ApplicationProvider {
	return td.icons
}

// SetIconProvider allows tests to set the icon provider used in this desktop
func (td *Desktop) SetIconProvider(icons fynedesk.ApplicationProvider) {
	td.icons = icons
}

// RecentApps returns applications that have recently been run. This test instance returns nil.
func (td *Desktop) RecentApps() []fynedesk.AppData {
	return nil
}

// Modules returns the list of modules currently loaded (by default no modules for this implementation)
func (*Desktop) Modules() []fynedesk.Module {
	return nil
}

// Root returns the root window, this is an in-memory test Fyne window
func (*Desktop) Root() fyne.Window {
	return test.NewWindow(nil)
}

// Run will run the desktop mainloop - no-op for testing
func (*Desktop) Run() {
}

// RunApp launches the passed application with appropriate environment setup
func (*Desktop) RunApp(app fynedesk.AppData) error {
	return app.Run([]string{}) // no added env
}

// Screens returns the list of screens this desktop runs on, by default a simple 2000x1000 value
func (td *Desktop) Screens() fynedesk.ScreenList {
	return td.screens
}

// Settings returns an in-memory test settings implementation
func (td *Desktop) Settings() fynedesk.DeskSettings {
	return td.settings
}

// ShowMenuAt is used to show a menu overlay above the desktop
func (td *Desktop) ShowMenuAt(menu *fyne.Menu, pos fyne.Position) {
	wid := widget.NewPopUpMenu(menu, test.Canvas())
	wid.Resize(wid.MinSize())
	wid.ShowAtPosition(pos)
}

// WindowManager returns the window manager for this desktop, an in-memory test instance unless
// configured through the constructor
func (td *Desktop) WindowManager() fynedesk.WindowManager {
	return td.wm
}
