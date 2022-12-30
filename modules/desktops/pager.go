package desktops

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
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

	p.ui = container.NewGridWithColumns(1, items...)
	return p
}
