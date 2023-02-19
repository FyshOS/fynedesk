package fyles

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Panel struct {
	widget.BaseWidget

	content *fyne.Container
	cb      func(fyne.URI)
	win     fyne.Window
	current *fileItem
}

func NewFylesPanel(c func(fyne.URI), w fyne.Window) *Panel {
	fileItemMin := fyne.NewSize(fileIconCellWidth, fileIconSize+fileTextSize+theme.InnerPadding())

	uiItems := container.NewGridWrap(fileItemMin)
	p := &Panel{content: uiItems, cb: c, win: w}
	p.ExtendBaseWidget(p)
	return p
}

func (p *Panel) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(p.content)
}

func (p *Panel) SetDir(u fyne.URI) {
	var items []fyne.CanvasObject
	parent, err := storage.Parent(u)
	if err == nil {
		up := &fileItem{parent: p, name: "(Parent)", location: parent, dir: true}
		up.ExtendBaseWidget(up)
		items = append(items, up)
	}
	list, err := storage.List(u)
	if err != nil {
		fyne.LogError("Could not read dir", err)
	} else {
		for _, item := range list {
			//if !ui.filter.Matches(item) {
			//	continue
			//}

			dir, _ := storage.CanList(item)
			items = append(items, newFileItem(item, dir, p))
		}
	}

	p.content.Objects = items
	p.content.Refresh()
}
