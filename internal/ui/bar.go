package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	deskDriver "fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fyshos.com/fynedesk"
	wmTheme "fyshos.com/fynedesk/theme"
)

// bar is the main widget housing app icons and taskbar area
type bar struct {
	widget.BaseWidget

	desk          fynedesk.Desktop    // The desktop instance we are holding icons for
	children      []fyne.CanvasObject // Icons that are laid out by the bar
	mouseInside   bool                // Is the mouse inside of the bar?
	mousePosition fyne.Position       // The current coordinates of the mouse cursor

	iconSize       float32
	iconScale      float32
	disableTaskbar bool
	disableZoom    bool
	icons          []*barIcon
	separator      *canvas.Rectangle
}

// MouseIn alerts the widget that the mouse has entered
func (b *bar) MouseIn(*deskDriver.MouseEvent) {
	if b.desk.Settings().LauncherDisableZoom() {
		return
	}
	b.mouseInside = true
	b.Refresh()
}

// MouseOut alerts the widget that the mouse has left
func (b *bar) MouseOut() {
	if b.desk.Settings().LauncherDisableZoom() {
		return
	}
	b.mouseInside = false
	b.Refresh()
}

// MouseMoved alerts the widget that the mouse has changed position
func (b *bar) MouseMoved(event *deskDriver.MouseEvent) {
	if b.desk.Settings().LauncherDisableZoom() {
		return
	}
	b.mousePosition = event.Position
	b.Refresh()
}

// append adds an object to the end of the widget
func (b *bar) append(object fyne.CanvasObject) {
	b.children = append(b.children, object)

	b.Refresh()
}

// appendSeparator adds a separator between the default icons and the taskbar
func (b *bar) appendSeparator() {
	b.separator = canvas.NewRectangle(theme.ForegroundColor())
	b.append(b.separator)
}

// removeFromTaskbar removes an object from the taskbar area of the widget
func (b *bar) removeFromTaskbar(object fyne.CanvasObject) {
	for i, icon := range b.children {
		if icon != object {
			continue
		}

		b.children = append(b.children[:i], b.children[i+1:]...)
		break
	}

	b.Refresh()
}

func (b *bar) newAppIcon(data fynedesk.AppData) *barIcon {
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

func (b *bar) createIcon(data fynedesk.AppData, win fynedesk.Window) *barIcon {
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

func (b *bar) taskbarIconTapped(win fynedesk.Window) {
	if win.Desktop() != fynedesk.Instance().Desktop() {
		b.desk.SetDesktop(win.Desktop())
		return
	}
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

func (b *bar) WindowAdded(win fynedesk.Window) {
	if win.Properties().SkipTaskbar() || b.desk.Settings().LauncherDisableTaskbar() {
		return
	}
	icon := b.createIcon(nil, win)
	if icon != nil {
		icon.onTapped = func() {
			b.taskbarIconTapped(win)
		}
		b.append(icon)
	}
}

func (b *bar) WindowRemoved(win fynedesk.Window) {
	if win.Properties().SkipTaskbar() || b.desk.Settings().LauncherDisableTaskbar() {
		return
	}
	for i, icon := range b.icons {
		if icon.windowData == nil || win != icon.windowData.win {
			continue
		}
		if !win.Iconic() {
			b.removeFromTaskbar(icon)
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
	b.disableTaskbar = disableTaskbar
	if disableTaskbar {
		return
	}
	b.appendSeparator()

	for _, win := range b.desk.WindowManager().Windows() {
		b.WindowAdded(win)
	}
}

func (b *bar) updateIconOrder() {
	var index = 0
	for i, obj := range b.children {
		if _, ok := obj.(*canvas.Rectangle); ok {
			index = i
			break
		}
	}
	var taskbarIcons []*barIcon
	if index != 0 {
		taskbarIcons = b.icons[index-1:]
	}

	b.icons = nil
	b.children = nil
	b.appendLauncherIcons()

	if b.desk.Settings().LauncherDisableTaskbar() {
		return
	}
	b.icons = append(b.icons, taskbarIcons...)
	for _, obj := range taskbarIcons {
		b.append(obj)
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

func (b *bar) appIcon(data fynedesk.AppData) fyne.Resource {
	return data.Icon(b.desk.Settings().IconTheme(), int((float32(b.iconSize)*b.iconScale)*b.desk.Screens().Primary().CanvasScale()))
}

func (b *bar) winIcon(win *appWindow) fyne.Resource {
	app := win.findApp()
	if app != nil {
		icon := b.appIcon(app)
		if icon != nil && icon != wmTheme.BrokenImageIcon {
			return icon
		}
	}

	iconRes := win.win.Properties().Icon()
	if iconRes == nil {
		return wmTheme.BrokenImageIcon
	}

	return iconRes
}

func (b *bar) appendLauncherIcons() {
	search := newBarIcon(theme.SearchIcon(), nil, nil)
	search.onTapped = ShowAppLauncher
	b.append(search)
	for _, name := range b.desk.Settings().LauncherIcons() {
		app := b.desk.IconProvider().FindAppFromName(name)
		if app == nil {
			continue
		}
		icon := b.createIcon(app, nil)
		if icon != nil {
			b.append(icon)
		}
	}
	if !b.desk.Settings().LauncherDisableTaskbar() {
		b.appendSeparator()
	}
}

// CreateRenderer creates the renderer that will be responsible for painting the widget
func (b *bar) CreateRenderer() fyne.WidgetRenderer {
	var bg fyne.CanvasObject
	if fynedesk.Instance().Settings().NarrowLeftLauncher() {
		bg = canvas.NewRectangle(wmTheme.WidgetPanelBackground())
	} else {
		bg = canvas.NewLinearGradient(theme.BackgroundColor(), color.Transparent, 180)
	}
	return &barRenderer{objects: b.children, background: bg, layout: newBarLayout(b), appBar: b}
}

// newBar creates a new application launcher and taskbar
func newBar(desk fynedesk.Desktop) *bar {
	bar := &bar{desk: desk}
	bar.ExtendBaseWidget(bar)
	bar.iconSize = float32(desk.Settings().LauncherIconSize())
	bar.iconScale = float32(desk.Settings().LauncherZoomScale())
	bar.disableTaskbar = desk.Settings().LauncherDisableTaskbar()

	if wm := desk.WindowManager(); wm != nil {
		wm.AddStackListener(bar)
	}
	bar.appendLauncherIcons()

	return bar
}

// barRenderer provides the renderer functions for the bar Widget
type barRenderer struct {
	layout barLayout

	appBar     *bar
	background fyne.CanvasObject
	objects    []fyne.CanvasObject
}

// MinSize returns the layout's Min Size
func (b *barRenderer) MinSize() fyne.Size {
	return b.layout.MinSize(b.objects)
}

// Layout recalculates the widget
func (b *barRenderer) Layout(size fyne.Size) {
	b.layout.setPointerInside(b.appBar.mouseInside)
	b.layout.setPointerPosition(b.appBar.mousePosition)
	b.layout.Layout(b.Objects(), size)
}

// BackgroundColor returns the background color of the widget
func (b *barRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

// Objects returns the objects associated with the widget
func (b *barRenderer) Objects() []fyne.CanvasObject {
	return append([]fyne.CanvasObject{b.background}, b.objects...)
}

// Refresh will recalculate the widget and repaint it
func (b *barRenderer) Refresh() {
	if fynedesk.Instance().Settings().NarrowLeftLauncher() {
		b.background = canvas.NewRectangle(wmTheme.WidgetPanelBackground())
	} else {
		b.background = canvas.NewLinearGradient(theme.BackgroundColor(), color.Transparent, 180)
	}
	if b.appBar.separator != nil {
		b.appBar.separator.FillColor = theme.ForegroundColor()
	}
	b.objects = b.appBar.children
	b.Layout(b.appBar.Size())

	canvas.Refresh(b.appBar.separator)
}

// Destroy tidies up resources
func (b *barRenderer) Destroy() {
}
