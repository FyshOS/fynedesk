package ui

import (
	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	"fyne.io/desktop"
)

func appExecPopUpListMatches(desk desktop.Desktop, win fyne.Window, appList *fyne.Container, input string) {
	iconTheme := desk.Settings().IconTheme()
	dataRange := desk.IconProvider().FindAppsMatching(input)
	for _, data := range dataRange {
		appData := data                     // capture for goroutine below
		icon := appData.Icon(iconTheme, 32) // TODO match theme but FDO needs power of 2 theme.IconInlineSize())
		app := widget.NewButtonWithIcon(appData.Name(), icon, func() {
			err := desk.RunApp(appData)
			if err != nil {
				fyne.LogError("Failed to start app", err)
				return
			}
			win.Close()
		})
		appList.AddObject(app)
	}
}

func newAppExecPopUp(desk desktop.Desktop) fyne.Window {
	win := fyne.CurrentApp().NewWindow("Applications")
	appList := fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	appScroller := widget.NewScrollContainer(appList)

	entry := widget.NewEntry()
	entry.SetPlaceHolder("Application")
	entry.OnChanged = func(input string) {
		appList.Objects = nil
		if input == "" {
			return
		}

		appExecPopUpListMatches(desk, win, appList, input)
	}

	cancel := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		win.Close()
	})

	content := fyne.NewContainerWithLayout(layout.NewBorderLayout(entry, cancel, nil, nil), entry, appScroller, cancel)

	win.SetContent(content)
	win.Resize(fyne.NewSize(300,
		cancel.MinSize().Height*4+theme.Padding()*6+entry.MinSize().Height))
	win.CenterOnScreen()
	win.Canvas().Focus(entry)
	return win
}
