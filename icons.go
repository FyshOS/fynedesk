package desktop

// IconData is an interface containing relavent information about application icons
type IconData interface {
	Name() string     //Name is the name of the app usually
	IconName() string //IconName is the name of the icon associated with an app
	IconPath() string //IconPath is the location of the app's icon
	Exec() string     //Exec is the command to run the app
}

// IconProvider describes a type that can locate icons and applications for the current system
type IconProvider interface {
	FindIconFromAppName(theme string, size int, appName string) IconData
	FindIconFromWinInfo(theme string, size int, win Window) IconData
	FindIconsMatchingAppName(theme string, size int, appName string) []IconData
}
