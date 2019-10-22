package desktop

import (
	"fyne.io/fyne"
)

var (
	appBar *bar
)

func barCreateIcon(b *bar, taskbar bool, data AppData, win Window) *barIcon {
	iconTheme := b.desk.Settings().IconTheme()
	if data == nil {
		return nil
	}
	iconRes := data.Icon(iconTheme, int(float32(b.iconSize)*b.iconScale))
	if iconRes == nil {
		return nil
	}
	icon := newBarIcon(iconRes)
	if taskbar == false {
		icon.onTapped = func() {
			err := b.desk.RunApp(data)
			if err != nil {
				fyne.LogError("Failed to start app", err)
			}
		}
	} else {
		icon.taskbarWindow = win
	}
	b.icons = append(b.icons, icon)
	return icon
}

func taskbarIconTapped(win Window) {
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

func (b *bar) WindowAdded(win Window) {
	if win.SkipTaskbar() {
		return
	}
	data := b.desk.IconProvider().FindAppFromWinInfo(win)
	if data == nil {
		return
	}
	icon := barCreateIcon(b, true, data, win)
	if icon != nil {
		icon.onTapped = func() {
			taskbarIconTapped(win)
		}
		appBar.append(icon)
	}
}

func (b *bar) WindowRemoved(win Window) {
	if win.SkipTaskbar() {
		return
	}
	for i, icon := range b.icons {
		if icon.taskbarWindow == nil || win != icon.taskbarWindow {
			continue
		}
		if !win.Iconic() {
			appBar.removeFromTaskbar(icon)
			b.icons = append(b.icons[:i], b.icons[i+1:]...)
		}
		break
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
