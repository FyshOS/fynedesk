package ui

import (
	"fyne.io/desktop"
	wmTheme "fyne.io/desktop/theme"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
)

var (
	appBar *bar
)

func (b *bar) newAppIcon(data desktop.AppData) *barIcon {
	iconRes := b.appIcon(data)
	icon := newBarIcon(iconRes, data, nil)

	icon.onTapped = func() {
		err := b.desk.RunApp(data)
		if err != nil {
			fyne.LogError("Failed to start app", err)
		}
	}

	return icon
}

func (b *bar) newTaskIcon(win *appWindow) *barIcon {
	iconRes := b.winIcon(win)
	return newBarIcon(iconRes, nil, win)
}

func (b *bar) createIcon(data desktop.AppData, win desktop.Window) *barIcon {
	if data == nil && win == nil {
		return nil
	}

	var icon *barIcon
	if win == nil {
		icon = b.newAppIcon(data)
	} else {
		icon = b.newTaskIcon(&appWindow{win: win, bar: b})
	}

	b.icons = append(b.icons, icon)
	return icon
}

func taskbarIconTapped(win desktop.Window) {
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

func (b *bar) WindowAdded(win desktop.Window) {
	if win.SkipTaskbar() || b.desk.Settings().LauncherDisableTaskbar() || win.Title() == "Fyne Desktop" {
		return
	}
	icon := b.createIcon(nil, win)
	if icon != nil {
		icon.onTapped = func() {
			taskbarIconTapped(win)
		}
		appBar.append(icon)
	}
}

func (b *bar) WindowRemoved(win desktop.Window) {
	if win.SkipTaskbar() || b.desk.Settings().LauncherDisableTaskbar() {
		return
	}
	for i, icon := range b.icons {
		if icon.windowData == nil || win != icon.windowData.win {
			continue
		}
		if !win.Iconic() {
			appBar.removeFromTaskbar(icon)
			b.icons = append(b.icons[:i], b.icons[i+1:]...)
		}
		break
	}
}

func (b *bar) updateTaskbar() {
	disableTaskbar := b.desk.Settings().LauncherDisableTaskbar()
	if disableTaskbar == b.disableTaskbar {
		return
	}
	b.disableTaskbar = b.desk.Settings().LauncherDisableTaskbar()
	if disableTaskbar == true {
		return
	}
	appBar.appendSeparator()
	for _, win := range b.desk.WindowManager().Windows() {
		b.WindowAdded(win)
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
	var taskbarIcons []*barIcon
	if index != -1 {
		taskbarIcons = b.icons[index:]
	}

	b.icons = nil
	b.children = nil
	b.appendLauncherIcons()

	if b.desk.Settings().LauncherDisableTaskbar() {
		return
	}
	b.icons = append(b.icons, taskbarIcons...)
	for _, obj := range taskbarIcons {
		appBar.append(obj)
	}
}

func (b *bar) updateIcons() {
	for _, icon := range b.icons {
		if icon.windowData != nil {
			icon.resource = b.winIcon(icon.windowData)
		} else {
			icon.resource = b.appIcon(icon.appData)
		}
		icon.Refresh()
	}
	b.Refresh()
}

func (b *bar) appIcon(data desktop.AppData) fyne.Resource {
	return data.Icon(b.desk.Settings().IconTheme(), int((float32(b.iconSize)*b.iconScale)*b.desk.Root().Canvas().Scale()))
}

func (b *bar) winIcon(win *appWindow) fyne.Resource {
	app := win.findApp()
	if app != nil {
		icon := b.appIcon(app)
		if icon != nil && icon != wmTheme.BrokenImageIcon {
			return icon
		}
	}

	iconRes := win.win.Icon()
	if iconRes == nil {
		return wmTheme.BrokenImageIcon
	}

	return iconRes
}

func (b *bar) appendLauncherIcons() {
	for _, name := range b.desk.Settings().LauncherIcons() {
		app := b.desk.IconProvider().FindAppFromName(name)
		if app == nil {
			continue
		}
		icon := appBar.createIcon(app, nil)
		if icon != nil {
			appBar.append(icon)
		}
	}
	if !b.desk.Settings().LauncherDisableTaskbar() {
		appBar.appendSeparator()
	}
}

func newBar(desk desktop.Desktop) fyne.CanvasObject {
	appBar = newAppBar(desk)

	if desk.WindowManager() != nil {
		desk.WindowManager().AddStackListener(appBar)
	}
	appBar.appendLauncherIcons()

	return appBar
}
