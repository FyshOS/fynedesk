package desktop

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne"
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

func barCreateIcon(taskbar bool, data IconData, win Window) *barIcon {
	if data == nil || data.IconPath() == "" {
		return nil
	}
	bytes, err := ioutil.ReadFile(data.IconPath())
	if err != nil {
		fyne.LogError("Could not read file", err)
		return nil
	}
	str := strings.Replace(data.IconPath(), "-", "", -1)
	iconResource := strings.Replace(str, "_", "", -1)

	res := fyne.NewStaticResource(strings.ToLower(filepath.Base(iconResource)), bytes)
	icon := newBarIcon(res)
	if taskbar == false {
		icon.onTapped = func() {
			command := strings.Split(data.Exec(), " ")
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
	data := GetIconDataByWinInfo(iconTheme, iconSize, win)
	if data == nil {
		return
	}
	icon := barCreateIcon(true, data, win)
	if icon != nil {
		icon.onTapped = func() {
			win.Focus()
		}
		appBar.append(icon)
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

	if wm != nil {
		bsl := &barStackListener{}
		wm.AddStackListener(bsl)
	}
	for _, app := range apps {
		data := GetIconDataByAppName(iconTheme, iconSize, app)
		if data == nil {
			continue
		}
		icon := barCreateIcon(false, data, nil)
		if icon != nil {
			appBar.append(icon)
		}
	}
	appBar.appendSeparator()

	return appBar
}
