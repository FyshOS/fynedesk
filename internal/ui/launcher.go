package ui

import (
	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	"fyne.io/desktop"
)

var appExec *launcher

type appEntry struct {
	widget.Entry

	launch *launcher
}

func (e *appEntry) TypedKey(ev *fyne.KeyEvent) {
	switch ev.Name {
	case fyne.KeyEscape:
		e.launch.close()
	case fyne.KeyReturn:
		e.launch.runSelected()
	case fyne.KeyUp:
		e.launch.setActiveIndex(e.launch.activeIndex - 1)
	case fyne.KeyDown:
		e.launch.setActiveIndex(e.launch.activeIndex + 1)
	default:
		e.Entry.TypedKey(ev)
	}
}

type launcher struct {
	win  fyne.Window
	desk desktop.Desktop

	entry       *appEntry
	appList     *fyne.Container
	activeIndex int
}

func (l *launcher) close() {
	l.win.Close()
}

func (l *launcher) runSelected() {
	if len(l.appList.Objects) == 0 {
		return
	}

	l.appList.Objects[l.activeIndex].(*widget.Button).OnTapped()
}

func (l *launcher) setActiveIndex(index int) {
	if index < 0 || index >= len(l.appList.Objects) {
		return
	}

	l.appList.Objects[l.activeIndex].(*widget.Button).Style = widget.DefaultButton
	l.appList.Objects[index].(*widget.Button).Style = widget.PrimaryButton
	l.activeIndex = index
	l.appList.Refresh()
}

func (l *launcher) runApp(app desktop.AppData) {
	err := l.desk.RunApp(app)
	if err != nil {
		fyne.LogError("Failed to start app", err)
		return
	}
	l.win.Close()
}

func (l *launcher) updateAppListMatching(input string) {
	l.activeIndex = 0
	l.appList.Objects = l.appButtonListMatching(input)
	l.appList.Refresh()
}

func (l *launcher) appButtonListMatching(input string) []fyne.CanvasObject {
	var appList []fyne.CanvasObject

	iconTheme := l.desk.Settings().IconTheme()
	dataRange := l.desk.IconProvider().FindAppsMatching(input)
	for i, data := range dataRange {
		appData := data // capture for goroutine below
		icon := appData.Icon(iconTheme, 32)
		app := widget.NewButtonWithIcon(appData.Name(), icon, func() {
			l.runApp(appData)
		})

		if i == 0 {
			app.Style = widget.PrimaryButton
		}
		appList = append(appList, app)
	}

	return appList
}

func newAppLauncher(desk desktop.Desktop) *launcher {
	win := fyne.CurrentApp().NewWindow("Applications")
	win.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		if ev.Name == fyne.KeyEscape {
			win.Close()
			return
		}
	})
	appList := fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	appScroller := widget.NewScrollContainer(appList)
	l := &launcher{win: win, desk: desk, appList: appList}

	entry := &appEntry{launch: l}
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

	content := fyne.NewContainerWithLayout(layout.NewBorderLayout(entry, cancel, nil, nil), entry, appScroller, cancel)

	win.SetContent(content)
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

	appExec = newAppLauncher(desktop.Instance())
	appExec.win.SetOnClosed(func() {
		appExec = nil
	})
	appExec.win.Show()
}
