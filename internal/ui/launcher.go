package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	deskDriver "fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fyshos.com/fynedesk"
	wmTheme "fyshos.com/fynedesk/theme"
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
	showMods bool

	entry       *appEntry
	appList     *fyne.Container
	appScroll   *container.Scroll
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

	oldActive := l.appList.Objects[l.activeIndex].(*widget.Button)
	oldActive.Importance = widget.MediumImportance
	oldActive.Refresh()
	active := l.appList.Objects[index].(*widget.Button)
	active.Importance = widget.HighImportance
	active.Refresh()

	l.activeIndex = index
	l.appScroll.Offset = fyne.NewPos(0,
		active.Position().Y+active.Size().Height/2-l.appScroll.Size().Height/2)
	l.appScroll.Refresh()
}

func (l *picker) updateAppListMatching(input string) {
	l.activeIndex = 0
	l.appScroll.ScrollToTop()
	l.appList.Objects = l.appButtonListMatching(input)
	l.appList.Refresh()
}

func (l *picker) appButtonListMatching(input string) []fyne.CanvasObject {
	var appList []fyne.CanvasObject
	var iconList = []fynedesk.AppData{}

	dataRange := l.desk.IconProvider().FindAppsMatching(input)
	for _, data := range dataRange {
		if data == nil || data.Hidden() {
			continue
		}
		appData := data // capture for goroutine below
		app := widget.NewButtonWithIcon(appData.Name(), wmTheme.BrokenImageIcon, func() {
			l.callback(appData)
			l.win.Close()
		})
		app.Alignment = widget.ButtonAlignLeading

		appList = append(appList, app)
		iconList = append(iconList, data)
	}
	go l.loadIcons(iconList, appList)

	appList = append(appList, l.loadSuggestionsMatching(input)...)
	if len(appList) > 0 {
		appList[0].(*widget.Button).Importance = widget.HighImportance
	}

	return appList
}

func (l *picker) loadIcons(dataRange []fynedesk.AppData, appList []fyne.CanvasObject) {
	iconTheme := l.desk.Settings().IconTheme()

	for i, data := range dataRange {
		app := appList[i].(*widget.Button)
		icon := data.Icon(iconTheme, 32)
		app.SetIcon(icon)
	}
}

func (l *picker) loadSuggestionsMatching(input string) []fyne.CanvasObject {
	var suggestList []fyne.CanvasObject

	for _, m := range l.desk.Modules() {
		suggest, ok := m.(fynedesk.LaunchSuggestionModule)
		if !ok {
			continue
		}

		for _, item := range suggest.LaunchSuggestions(input) {
			launchData := item // capture for goroutine below
			button := widget.NewButtonWithIcon(item.Title(), item.Icon(), func() {
				l.win.Close()
				launchData.Launch()
			})

			suggestList = append(suggestList, button)
		}
	}

	return suggestList
}

func (l *picker) Show() {
	l.win.Show()
}

func newAppPicker(title string, callback func(fynedesk.AppData)) *picker {
	var win fyne.Window
	if d, ok := fyne.CurrentApp().Driver().(deskDriver.Driver); ok {
		win = d.CreateSplashWindow()
		win.SetPadded(true)
		win.SetTitle(title)
	} else {
		win = fyne.CurrentApp().NewWindow(title)
	}

	win.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		if ev.Name == fyne.KeyEscape {
			win.Close()
			return
		}
	})

	appList := container.NewVBox()
	appScroller := container.NewScroll(appList)
	l := &picker{win: win, desk: fynedesk.Instance(), appList: appList, appScroll: appScroller, callback: callback}

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

	win.SetContent(container.NewBorder(entry, cancel, nil, nil, appScroller))
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
		return
	}

	appExec = newAppPicker("Application Launcher "+SkipTaskbarHint, func(app fynedesk.AppData) {
		err := fynedesk.Instance().RunApp(app)
		if err != nil {
			fyne.LogError("Failed to start app", err)
			return
		}
	})
	appExec.showMods = true
	appExec.win.SetOnClosed(func() {
		appExec = nil
	})
	appExec.Show()
}
