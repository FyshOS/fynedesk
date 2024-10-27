package fynedesk

import "fyne.io/fyne/v2"

// AppData is an interface for accessing information about application icons
type AppData interface {
	Name() string       // Name is the name of the app usually
	Run([]string) error // Run is the command to run the app, passing any environment variables to be set

	Categories() []string                      // Categories is a list of categories that the app fits in (platform specific)
	Hidden() bool                              // Hidden specifies whether instances of this app should be hidden
	Icon(theme string, size int) fyne.Resource // Icon returns an icon for the app in the requested theme and size

	Source() *AppSource // Source will return the location of the app source code from metadata, if known
}

// AppSource represents the source code informtion of an application
type AppSource struct {
	Repo, Dir string
}

// ApplicationProvider describes a type that can locate icons and applications for the current system
type ApplicationProvider interface {
	AvailableApps() []AppData
	AvailableThemes() []string
	FindAppFromName(appName string) AppData
	FindAppFromWinInfo(win Window) AppData
	FindAppsMatching(pattern string) []AppData
	DefaultApps() []AppData
	CategorizedApps() map[string][]AppData
}
