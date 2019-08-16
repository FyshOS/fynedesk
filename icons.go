package desktop

import (
	"github.com/fyne-io/desktop/internal"
)

//IconData is an interface containing relavent information about application icons
type IconData interface {
	Name() string     //Name is the name of the app usually
	IconName() string //IconName is the name of the icon associated with an app
	IconPath() string //IconPath is the location of the app's icon
	Exec() string     //Exec is the command to run the app
}

//GetIconDataByAppName matches an icon name to a location and returns an IconData interface
func GetIconDataByAppName(theme string, size int, appName string) IconData {
	fdoIcon := internal.FdoLookupApplication(theme, size, appName)
	return fdoIcon
}

//GetIconDataByWinInfo matches window information to an icon location and returns an IconData interface
func GetIconDataByWinInfo(theme string, size int, win Window) IconData {
	fdoIcon := internal.FdoLookupApplicationWinInfo(theme, size, win.Title(), win.Class(), win.Command(), win.IconName())
	if fdoIcon == nil {
		return nil
	}
	return fdoIcon
}
