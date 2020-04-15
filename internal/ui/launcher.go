package ui

import (
	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	"fyne.io/fynedesk"
)

var appExec *picker

type appEntry struct {
	widget.Entry

	pick *picker
}

func (e *appEntry) TypedKey(ev *fyne.KeyEvent) {
	switch ev.Name {
	case fyne.KeyEscape:
		e.pick.close()
	case fyne.KeyReturn:
		e.pick.pickSelected()
	case fyne.KeyUp:
		e.pick.setActiveIndex(e.pick.activeIndex - 1)
	case fyne.KeyDown:
		e.pick.setActiveIndex(e.pick.activeIndex + 1)
	default:
		e.Entry.TypedKey(ev)
	}
}

type picker struct {
	win      fyne.Window
	desk     fynedesk.Desktop
	callback func(data fynedesk.AppData)

	entry       *appEntry
	appList     *fyne.Container
	activeIndex int
}

func (l *picker) close() {
	l.win.Close()
}

func (l *picker) pickSelected() {
	if len(l.appList.Objects) == 0 {
		return
	}

	l.appList.Objects[l.activeIndex].(*widget.Button).OnTapped()
}

func (l *picker) setActiveIndex(index int) {
	if index < 0 || index >= len(l.appList.Objects) {
		return
	}

	l.appList.Objects[l.activeIndex].(*widget.Button).Style = widget.DefaultButton
	l.appList.Objects[index].(*widget.Button).Style = widget.PrimaryButton
	l.activeIndex = index
	l.appList.Refresh()
}

func (l *picker) updateAppListMatching(input string) {
	l.activeIndex = 0
	l.appList.Objects = l.appButtonListMatching(input)
	l.appList.Refresh()
}

func (l *picker) appButtonListMatching(input string) []fyne.CanvasObject {
	var appList []fyne.CanvasObject

	iconTheme := l.desk.Settings().IconTheme()
	dataRange := l.desk.IconProvider().FindAppsMatching(input)
	for i, data := range dataRange {
		appData := data // capture for goroutine below
		icon := appData.Icon(iconTheme, 32)
		app := widget.NewButtonWithIcon(appData.Name(), icon, func() {
			l.callback(appData)
			l.win.Close()
		})

		if i == 0 {
			app.Style = widget.PrimaryButton
		}
		appList = append(appList, app)
	}

	return appList
}

func (l *picker) Show() {
	l.win.Show()
}

func newAppPicker(title string, callback func(fynedesk.AppData)) *picker {
	win := fyne.CurrentApp().NewWindow(title)
	win.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		if ev.Name == fyne.KeyEscape {
			win.Close()
			return
		}
	})
	appList := fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	appScroller := widget.NewScrollContainer(appList)
	l := &picker{win: win, desk: fynedesk.Instance(), appList: appList, callback: callback}

	entry := &appEntry{pick: l}
	entry.ExtendBaseWidget(entry)
	entry.SetPlaceHolder("Application")
	entry.OnChanged = func(input string) {
		appList.Objects = nil
		if input == "" {
			return
		}

		l.updateAppListMatching(input)
	}
	l.entry = entry

	cancel := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		win.Close()
	})

	win.SetContent(fyne.NewContainerWithLayout(layout.NewBorderLayout(entry, cancel, nil, nil),
		entry, appScroller, cancel))
	win.Resize(fyne.NewSize(300,
		cancel.MinSize().Height*4+theme.Padding()*6+entry.MinSize().Height))
	win.CenterOnScreen()
	win.Canvas().Focus(entry)
	return l
}

// ShowAppLauncher opens a new application launcher, closing an old one if it existed.
func ShowAppLauncher() {
	if appExec != nil {
		appExec.close()
	}

	appExec = newAppPicker("Application Launcher", func(app fynedesk.AppData) {
		err := fynedesk.Instance().RunApp(app)
		if err != nil {
			fyne.LogError("Failed to start app", err)
			return
		}
	})
	appExec.win.SetOnClosed(func() {
		appExec = nil
	})
	appExec.Show()
}
