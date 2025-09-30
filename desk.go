package fynedesk // import "fyshos.com/fynedesk"

import (
	"fyne.io/fyne/v2"

	"github.com/FyshOS/appie"
)

// Desktop defines an embedded or full desktop environment that we can run.
type Desktop interface {
	Run()
	RunApp(appie.AppData) error
	RecentApps() []appie.AppData
	Settings() DeskSettings
	ContentBoundsPixels(*Screen) (x, y, w, h uint32)
	RootSizePixels() (w, h uint32)
	Screens() ScreenList

	IconProvider() appie.Provider
	WindowManager() WindowManager
	Modules() []Module

	AddShortcut(shortcut *Shortcut, handler func())
	ShowMenuAt(menu *fyne.Menu, pos fyne.Position)
	Root() fyne.Window

	Desktop() int
	SetDesktop(int)
	ShowSettings()

	DelayScreenSaver()
	TriggerScreenSaver()
}

var instance Desktop

// Instance returns the current desktop environment and provides access to injected functionality.
func Instance() Desktop {
	return instance
}

// SetInstance is an internal call :( TODO
func SetInstance(desk Desktop) {
	instance = desk
}
