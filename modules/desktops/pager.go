package desktops

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fyshos.com/tyde"
)

// pagerMaxCols is how many desktops sit side by side in the wide widget panel.
const pagerMaxCols = 4

type pager struct {
	buttons, labels *fyne.Container
	wins            *fyne.Container
	winObjs         map[tyde.Window]fyne.CanvasObject // quick lookup for move updates
}

func newPager(d *desktops) *pager {
	p := &pager{wins: container.NewWithoutLayout()}

	count := deskCount()
	buttons := make([]fyne.CanvasObject, count)
	labels := make([]fyne.CanvasObject, count)
	for i := 0; i < count; i++ {
		id := strconv.Itoa(i + 1)
		deskID := i
		buttons[i] = widget.NewButton("", func() {
			d.setDesktop(deskID)
		})
		labels[i] = widget.NewLabelWithStyle(id, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	}

	cols := pagerMaxCols
	if count < cols {
		cols = count
	}
	if tyde.Instance().Settings().NarrowWidgetPanel() {
		cols = 1
	}
	p.buttons = container.NewGridWithColumns(cols, buttons...)
	p.labels = container.NewGridWithColumns(cols, labels...)
	p.refresh()
	tyde.Instance().WindowManager().AddStackListener(p)

	return p
}

func (p *pager) WindowAdded(_ tyde.Window) {
	p.refresh()
}

func (p *pager) WindowMoved(win tyde.Window) {
	obj, ok := p.winObjs[win]
	if !ok {
		p.refresh()
		return
	}

	fyne.Do(func() {
		if !p.place(win, obj) {
			obj.Hide()
		}
		p.wins.Refresh()
	})
}

func (p *pager) WindowOrderChanged() {
	p.refresh()
}

func (p *pager) WindowRemoved(_ tyde.Window) {
	p.refresh()
}

func (p *pager) WindowStateChanged(_ tyde.Window) {
}

func (p *pager) refresh() {
	fyne.Do(p.refreshNow)
}

func (p *pager) refreshButtons() {
	desk := tyde.Instance()
	fyne.Do(func() {
		for i, b := range p.buttons.Objects {
			l := p.labels.Objects[i]
			if i == desk.Desktop() {
				b.(*widget.Button).Importance = widget.HighImportance
				l.(*widget.Label).Importance = widget.LowImportance
			} else {
				b.(*widget.Button).Importance = widget.MediumImportance
				l.(*widget.Label).Importance = widget.MediumImportance
			}

			b.Refresh()
			l.Refresh()
		}
	})
}

func (p *pager) refreshNow() {
	wins := tyde.Instance().WindowManager().Windows()

	p.refreshButtons()

	var rects []fyne.CanvasObject
	winObjs := make(map[tyde.Window]fyne.CanvasObject, len(wins))

	for j := len(wins) - 1; j >= 0; j-- {
		win := wins[j]
		if win.Iconic() || win.Properties().SkipTaskbar() {
			continue
		}

		var obj fyne.CanvasObject
		obj = canvas.NewRectangle(theme.Color(theme.ColorNameDisabled))
		if win.Properties().Icon() != nil {
			obj = container.NewStack(obj,
				canvas.NewImageFromResource(win.Properties().Icon()))
		}
		if !p.place(win, obj) {
			continue
		}
		rects = append(rects, obj)
		winObjs[win] = obj
	}

	p.winObjs = winObjs
	p.wins.Objects = rects
	p.wins.Refresh()
}

// place sizes and positions obj, the preview of win, within the pager cell of the
// desktop the window is on, scaled from the window's screen to the cell. Window
// positions are relative to the current desktop, so the page offset of the window's
// own desktop is removed to find where it sits on that desktop. A pinned window is
// shown on the current desktop. Returns false if the window's desktop has no cell.
func (p *pager) place(win tyde.Window, obj fyne.CanvasObject) bool {
	desk := tyde.Instance()
	current := desk.Desktop()
	id := win.Desktop()
	if win.Pinned() {
		id = current
	}
	if id < 0 || id >= len(p.buttons.Objects) {
		return false
	}

	cell := p.buttons.Objects[id]
	screen := desk.Screens().ScreenForWindow(win)
	scale := screen.CanvasScale()
	_, rootHeight := desk.RootSizePixels()
	pageY := float32((id - current) * int(rootHeight))

	x := (win.Position().X*scale - float32(screen.X)) / float32(screen.Width) * cell.Size().Width
	y := (win.Position().Y*scale - float32(screen.Y) - pageY) / float32(screen.Height) * cell.Size().Height
	w := win.Size().Width * scale / float32(screen.Width) * cell.Size().Width
	h := win.Size().Height * scale / float32(screen.Height) * cell.Size().Height
	obj.Resize(fyne.NewSize(w, h))
	obj.Move(cell.Position().Add(fyne.NewPos(x, y)))
	return true
}
