package desktops

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fyshos.com/fynedesk"
)

type pager struct {
	ui   fyne.CanvasObject
	wins *fyne.Container
}

func newPager(d *desktops) *pager {
	p := &pager{wins: container.NewWithoutLayout()}

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
	fynedesk.Instance().WindowManager().AddStackListener(p)

	return p
}

func (p *pager) WindowAdded(_ fynedesk.Window) {
	p.refresh()
}

func (p *pager) WindowMoved(_ fynedesk.Window) {
	p.refresh()
}

func (p *pager) WindowOrderChanged() {
	p.refresh()
}

func (p *pager) WindowRemoved(_ fynedesk.Window) {
	p.refresh()
}

func (p *pager) refresh() {
	desk := fynedesk.Instance()
	p.refreshFrom(desk.Desktop())
}

func (p *pager) refreshFrom(oldID int) {
	desk := fynedesk.Instance()
	wins := fynedesk.Instance().WindowManager().Windows()

	var rects []fyne.CanvasObject
	for i, b := range p.ui.(*fyne.Container).Objects {
		if i == desk.Desktop() {
			b.(*widget.Button).Importance = widget.HighImportance
		} else {
			b.(*widget.Button).Importance = widget.MediumImportance
		}
		b.Refresh()
	}
	pivot := p.ui.(*fyne.Container).Objects[oldID]

	for j := len(wins) - 1; j >= 0; j-- {
		win := wins[j]
		if win.Iconic() {
			continue
		}

		yPad := theme.Padding() * float32(win.Desktop()-oldID)
		screen := fynedesk.Instance().Screens().ScreenForWindow(win)

		var obj fyne.CanvasObject
		obj = canvas.NewRectangle(theme.DisabledColor())
		if win.Properties().Icon() != nil {
			obj = container.NewStack(obj,
				canvas.NewImageFromResource(win.Properties().Icon()))
		}
		rects = append(rects, obj)

		x := (win.Position().X * screen.Scale) / float32(screen.Width) * pivot.Size().Width
		y := (win.Position().Y * screen.Scale) / float32(screen.Height) * pivot.Size().Height
		w := (win.Size().Width * screen.Scale) / float32(screen.Width) * pivot.Size().Width
		h := (win.Size().Height * screen.Scale) / float32(screen.Height) * pivot.Size().Height
		obj.Resize(fyne.NewSize(w, h))
		obj.Move(pivot.Position().Add(fyne.NewPos(x, y+yPad)))
	}

	p.wins.Objects = rects
	p.wins.Refresh()
}
