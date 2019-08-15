package desktop

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
)

var (
	appBar    *bar
	apps      = []string{"xterm", "gimp", "google-chrome", "firefox"}
	iconSize  = 32
	iconTheme = "hicolor"
	icons     []*barIcon
)

type barStackListener struct {
}

func barCreateIcon(taskbar bool, idata *IconData, win Window) *barIcon {
	if idata == nil || idata.IconPath == "" {
		return nil
	}
	bytes, err := ioutil.ReadFile(idata.IconPath)
	if err != nil {
		fyne.LogError("Could not read file", err)
		return nil
	}
	str := strings.Replace(idata.IconPath, "-", "", -1)
	iconResource := strings.Replace(str, "_", "", -1)

	res := fyne.NewStaticResource(strings.ToLower(filepath.Base(iconResource)), bytes)
	icon := newBarIcon(res)
	if taskbar == false {
		icon.onTapped = func() {
			command := strings.Split(idata.Exec, " ")
			if len(command) > 1 {
				args := fmt.Sprintf("%s", command[1:])
				exec.Command(command[0], args).Start()
			} else {
				exec.Command(command[0]).Start()
			}
		}
	} else {
		icon.taskbarWindow = win
	}
	icons = append(icons, icon)
	return icon
}

func (bsl *barStackListener) WindowAdded(win Window) {
	fdoDesktop := FdoLookupApplicationWinInfo(win.Title(), win.Class(), win.Command(), win.IconName())
	icon := barCreateIcon(true, fdoDesktop, win)
	if icon != nil {
		icon.onTapped = func() {
			win.Focus()
		}
		appBar.appendToTaskbar(icon)
	}
}

func (bsl *barStackListener) WindowRemoved(win Window) {
	for i, icon := range icons {
		if icon.taskbarWindow != nil {
			if win == icon.taskbarWindow {
				appBar.removeFromTaskbar(icon)
				icons = append(icons[:i], icons[i+1:]...)
			}
		}
	}
}

func newBar(wm WindowManager) fyne.CanvasObject {
	appBar = newAppBar()
	appBar.append(layout.NewSpacer())

	if wm != nil {
		bsl := &barStackListener{}
		wm.AddStackListener(bsl)
	}
	for _, app := range apps {
		fdoDesktop := FdoLookupApplication(app)
		icon := barCreateIcon(false, fdoDesktop, nil)
		if icon != nil {
			appBar.append(icon)
		}
	}
	appBar.appendSeparator()
	appBar.append(layout.NewSpacer())

	return appBar
}
