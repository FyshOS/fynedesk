package icon

import (
	"runtime"

	"fyshos.com/fynedesk"
	"github.com/FyshOS/appie"
)

// FindAppFromWinInfo searches the known applications and tries to find one
// based on the properties of an open window.
func FindAppFromWinInfo(win fynedesk.Window, provider appie.Provider) appie.AppData {
	if runtime.GOOS == "darwin" { // simpler handling when we are not the desktop environment
		title := win.Properties().Title()
		if title != "" {
			apps := provider.FindAppsMatching(title)
			if len(apps) > 0 {
				return apps[0]
			}
		}
		return nil
	}

	cmd := win.Properties().Command()
	if cmd != "" {
		apps := provider.FindAppsMatching(cmd)
		if len(apps) > 0 {
			return apps[0]
		}
	}

	for _, class := range win.Properties().Class() {
		if class != "" {
			apps := provider.FindAppsMatching(class)
			if len(apps) > 0 {
				return apps[0]
			}
		}
	}

	icon := win.Properties().IconName()
	if icon != "" {
		apps := provider.FindAppsMatching(icon)
		if len(apps) > 0 {
			return apps[0]
		}
	}
	return nil
}
