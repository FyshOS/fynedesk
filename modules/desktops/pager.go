package desktops

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fynedesk"
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
	return p
}
