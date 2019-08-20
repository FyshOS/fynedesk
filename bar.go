package desktop

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne"
)

var (
	appBar   *bar
	apps     = []string{"xterm", "gimp", "google-chrome", "firefox"}
	iconSize = 32
	icons    []*barIcon
)

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
			exec.Command(command[0]).Start()
		}
	} else {
		icon.taskbarWindow = win
	}
	icons = append(icons, icon)
	return icon
}

func (b *bar) WindowAdded(win Window) {
	iconTheme := b.desk.Settings().IconTheme()
	data := b.desk.IconProvider().FindIconFromWinInfo(iconTheme, iconSize, win)
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

func (b *bar) WindowRemoved(win Window) {
	for i, icon := range icons {
		if icon.taskbarWindow != nil {
			if win == icon.taskbarWindow {
				appBar.removeFromTaskbar(icon)
				icons = append(icons[:i], icons[i+1:]...)
			}
		}
	}
}

func newBar(desk Desktop) fyne.CanvasObject {
	appBar = newAppBar(desk)

	if desk.WindowManager() != nil {
		desk.WindowManager().AddStackListener(appBar)
	}
	for _, app := range apps {
		iconTheme := desk.Settings().IconTheme()
		data := desk.IconProvider().FindIconFromAppName(iconTheme, iconSize, app)
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
