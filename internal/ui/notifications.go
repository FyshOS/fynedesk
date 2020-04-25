package ui

import (
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"

	"fyne.io/fynedesk/wm"
)

type notification struct {
	message *wm.Notification

	renderer fyne.CanvasObject
}

func (n *notification) show(list *fyne.Container) {
	title := widget.NewLabel(n.message.Title)
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Wrapping = fyne.TextTruncate
	text := widget.NewLabel(n.message.Body)
	text.Wrapping = fyne.TextWrapWord
	n.renderer = widget.NewVBox(title, text)

	list.Objects = append(list.Objects, n.renderer)
	list.Refresh()

	go func() {
		time.Sleep(time.Second * 10)

		n.hide(list)
	}()
}

func (n *notification) hide(list *fyne.Container) {
	var items []fyne.CanvasObject
	for _, item := range list.Objects {
		if item == n.renderer {
			continue
		}

		items = append(items, item)
	}

	list.Objects = items
	list.Refresh()
}

type notifications struct {
	list *fyne.Container
}

func (n *notifications) newMessage(message *wm.Notification) {
	item := &notification{message: message}
	go item.show(n.list)
}

func startNotifications() fyne.CanvasObject {
	box := fyne.NewContainerWithLayout(layout.NewVBoxLayout())

	n := &notifications{list: box}
	wm.SetNotificationListener(n.newMessage)

	return box
}
