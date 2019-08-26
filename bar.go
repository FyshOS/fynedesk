package desktop

import (
	"fyne.io/fyne"
)

var (
	appBar    *bar
	iconSize  = 32
	iconScale = 2.0
	icons     []*barIcon
)

func barCreateIcon(b *bar, taskbar bool, data AppData, win Window) *barIcon {
	iconTheme := b.desk.Settings().IconTheme()
	if data == nil {
		return nil
	}
	iconRes := data.Icon(iconTheme, int(float64(iconSize)*iconScale))
	if iconRes == nil {
		return nil
	}
	icon := newBarIcon(iconRes)
	if taskbar == false {
		icon.onTapped = func() {
			err := data.Run()
			if err != nil {
				fyne.LogError("Failed to start app", err)
			}
		}
	} else {
		icon.taskbarWindow = win
	}
	icons = append(icons, icon)
	return icon
}

func (b *bar) WindowAdded(win Window) {
	data := b.desk.IconProvider().FindAppFromWinInfo(win)
	if data == nil {
		return
	}
	icon := barCreateIcon(b, true, data, win)
	if icon != nil {
		icon.onTapped = func() {
			if !win.Iconic() && win.TopWindow() {
				win.Iconify()
				return
			}
			if win.Iconic() {
				win.Uniconify()
			}
			win.RaiseToTop()
			win.Focus()
		}
		appBar.append(icon)
	}
}

func (b *bar) WindowRemoved(win Window) {
	for i, icon := range icons {
		if icon.taskbarWindow != nil {
			if win == icon.taskbarWindow {
				if !win.Iconic() {
					appBar.removeFromTaskbar(icon)
					icons = append(icons[:i], icons[i+1:]...)
				}
			}
		}
	}
}

func newBar(desk Desktop) fyne.CanvasObject {
	appBar = newAppBar(desk)

	if desk.WindowManager() != nil {
		desk.WindowManager().AddStackListener(appBar)
	}
	for _, app := range appBar.desk.IconProvider().DefaultApps() {
		icon := barCreateIcon(appBar, false, app, nil)
		if icon != nil {
			appBar.append(icon)
		}
	}
	appBar.appendSeparator()

	return appBar
}
