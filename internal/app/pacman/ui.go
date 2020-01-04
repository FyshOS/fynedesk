package pacman

import (
	"fmt"

	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

type updater struct {
	manager *Manager

	win         fyne.Window
	description *widget.Label
}

func (u *updater) loadUpdates(list *widget.Box, status *widget.Label) {
	if u.manager == nil || !u.manager.HasUpdates() {
		list.Append(widget.NewLabel("No Updates Found"))
		status.SetText("0 Updates")
		return
	}

	updates := u.manager.Updates()
	for _, update := range updates {
		list.Append(widget.NewLabel(update.Name))
	}
	status.SetText(fmt.Sprintf("%d Updates Available", len(updates)))
}

func (u *updater) loadUpdateUI() fyne.CanvasObject {
	list := widget.NewVBox()
	status := widget.NewLabel("Status")
	go u.loadUpdates(list, status)

	update := widget.NewButton("Update", func() {
		progress := dialog.NewProgress("Updating", "Updating packages", u.win)
		go func() {
			progress.Show()
			u.manager.UpdateAll()
			progress.Hide()
		}()
	})
	buttons := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, nil, update),
		update, status)
	return fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, buttons, nil, nil),
		buttons, widget.NewScrollContainer(list))
}

func (u *updater) loadUI(index int) fyne.CanvasObject {
	tabs := widget.NewTabContainer(
		widget.NewTabItem("Search", u.loadSearchUI()),
		widget.NewTabItem("Update", u.loadUpdateUI()))

	tabs.SelectTabIndex(index)
	return tabs
}

func runAtTab(tab int) *updater {
	m, err := NewManager()
	win := fyne.CurrentApp().NewWindow("Package Manager")

	update := &updater{manager: m, win: win}
	ui := update.loadUI(tab)

	win.SetContent(ui)
	win.Resize(fyne.NewSize(360, 280))
	win.Show()

	if err != nil {
		dialog.ShowError(err, win)
	}

	return update
}

// ShowManage opens a new app manager window with the package manager search screen open
func ShowManage() {
	runAtTab(0)
}

// ShowUpdate opens a new app manager window with the updates screen open
func ShowUpdate() {
	u := runAtTab(1)
	if u.manager.provider == nil {
		return
	}

	if !u.manager.HasUpdates() {
		dialog.ShowInformation("System up to date", "There are no updates available", u.win)
	}
}
