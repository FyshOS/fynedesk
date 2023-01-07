package desktops

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fyshos.com/fynedesk"
)

type pager struct {
	ui fyne.CanvasObject
}

func newPager(d *desktops) *pager {
	p := &pager{}

	items := make([]fyne.CanvasObject, deskCount)
	for i := 0; i < deskCount; i++ {
		id := strconv.Itoa(i + 1)
		deskID := i
		items[i] = widget.NewButton(id, func() {
			d.setDesktop(deskID)
		})
	}

	if fynedesk.Instance() != nil && fynedesk.Instance().Settings().NarrowWidgetPanel() {
		p.ui = container.NewGridWithColumns(1, items...)
	} else {
		p.ui = container.NewGridWithColumns(4, items...)
	}
	p.refresh()

	return p
}

func (p *pager) refresh() {
	desk := fynedesk.Instance()

	for i, b := range p.ui.(*fyne.Container).Objects {
		if i == desk.Desktop() {
			b.(*widget.Button).Importance = widget.HighImportance
		} else {
			b.(*widget.Button).Importance = widget.MediumImportance
		}
		b.Refresh()
	}
}
