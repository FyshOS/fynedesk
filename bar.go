package desktop

import (
	wmTheme "fyne.io/desktop/theme"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
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
	if iconRes == nil || iconRes == wmTheme.BrokenImageIcon {
		if win != nil {
			iconRes = win.Icon()
			if iconRes == nil {
				iconRes = wmTheme.BrokenImageIcon
			}
		}
	}
	icon := newBarIcon(iconRes, data)
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

func (b *bar) updateIconOrder() {
	var index = -1
	for i, obj := range b.children {
		if _, ok := obj.(*canvas.Rectangle); ok {
			index = i
			break
		}
	}
	if index == -1 {
		return
	}
	var taskbarIcons = b.icons[index:]

	b.icons = nil
	b.children = nil
	b.appendDefaultIcons()

	b.icons = append(b.icons, taskbarIcons...)
	for _, obj := range taskbarIcons {
		appBar.append(obj)
	}
}

func (b *bar) updateIcons() {
	iconTheme := b.desk.Settings().IconTheme()
	for _, icon := range b.icons {
		iconRes := icon.appData.Icon(iconTheme, int(float32(b.iconSize)*b.iconScale))
		if iconRes == nil || iconRes == wmTheme.BrokenImageIcon {
			if icon.taskbarWindow != nil {
				iconRes = icon.taskbarWindow.Icon()
				if iconRes == nil {
					iconRes = wmTheme.BrokenImageIcon
				}
			}
		}
		icon.resource = iconRes
		icon.Refresh()
	}
}

func (b *bar) appendDefaultIcons() {
	for _, name := range b.desk.Settings().DefaultApps() {
		app := b.desk.IconProvider().FindAppFromName(name)
		if app == nil {
			continue
		}
		icon := barCreateIcon(appBar, false, app, nil)
		if icon != nil {
			appBar.append(icon)
		}
	}
	appBar.appendSeparator()
}

func newBar(desk Desktop) fyne.CanvasObject {
	appBar = newAppBar(desk)

	if desk.WindowManager() != nil {
		desk.WindowManager().AddStackListener(appBar)
	}
	appBar.appendDefaultIcons()

	return appBar
}
