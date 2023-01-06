package ui

import (
	"image/color"
	"os"
	"path"
	"time"

	"github.com/disintegration/imaging"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/software"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fyshos.com/fynedesk"
	wmtheme "fyshos.com/fynedesk/theme"
)

type widgetRenderer struct {
	panel *widgetPanel
	bg    *canvas.Rectangle

	layout  fyne.Layout
	objects []fyne.CanvasObject
}

func (w *widgetRenderer) MinSize() fyne.Size {
	return w.layout.MinSize(w.objects)
}

func (w *widgetRenderer) Layout(size fyne.Size) {
	w.bg.Resize(size)
	w.layout.Layout(w.objects[1:], size)
}

func (w *widgetRenderer) Refresh() {
	w.bg.FillColor = wmtheme.WidgetPanelBackground()
	w.bg.Refresh()

	w.panel.account.SetText(w.panel.accountLabel())
	if w.panel.desk.Settings().NarrowWidgetPanel() {
		w.panel.search.Hide()
		w.panel.clocks.Objects[0].Hide()
		w.panel.clocks.Objects[1].Show()
	} else {
		w.panel.search.Show()
		w.panel.clocks.Objects[0].Show()
		w.panel.clocks.Objects[1].Hide()
	}
	w.panel.clock.Color = theme.ForegroundColor()
	w.panel.vClock.Color = theme.ForegroundColor()
	canvas.Refresh(w.panel.clock)
}

func (w *widgetRenderer) Objects() []fyne.CanvasObject {
	return w.objects
}

func (w *widgetRenderer) Destroy() {
}

type widgetPanel struct {
	widget.BaseWidget

	desk            fynedesk.Desktop
	about, settings fyne.Window

	account, search *widget.Button
	clock, vClock   *canvas.Text
	date            *widget.Label
	rotated         *canvas.Image
	modules, clocks *fyne.Container
	notifications   fyne.CanvasObject
}

func (w *widgetPanel) clockTick() {
	tick := time.NewTicker(time.Second)
	go func() {
		for {
			<-tick.C
			w.clock.Text = w.formattedTime()
			w.vClock.Text = w.formattedTime()
			canvas.Refresh(w.clock)
			if w.desk.Settings().NarrowWidgetPanel() {
				w.rotate(w.vClock)
			}

			w.date.SetText(w.formattedDate())
			canvas.Refresh(w.date)
		}
	}()
}

func (w *widgetPanel) formattedTime() string {
	if w.desk.Settings().ClockFormatting() == "12h" {
		return time.Now().Format("03:04pm")
	}

	return time.Now().Format("15:04")
}

func (w *widgetPanel) formattedDate() string {
	format := "2 Jan"
	if w.desk.Settings().NarrowWidgetPanel() {
		format = "2\nJan"
	}
	return time.Now().Format(format)
}

func (w *widgetPanel) createClock() {
	var style fyne.TextStyle
	style.Monospace = true

	w.clock = &canvas.Text{
		Color:     theme.ForegroundColor(),
		Text:      w.formattedTime(),
		Alignment: fyne.TextAlignCenter,
		TextStyle: style,
		TextSize:  3 * theme.TextSize(),
	}
	w.vClock = &canvas.Text{
		Color:     theme.ForegroundColor(),
		Text:      w.formattedTime(),
		Alignment: fyne.TextAlignCenter,
		TextStyle: style,
		TextSize:  wmtheme.NarrowBarWidth * 1.5,
	}
	w.date = &widget.Label{
		Text:      w.formattedDate(),
		Alignment: fyne.TextAlignCenter,
		TextStyle: style,
	}

	go w.clockTick()
}

func (w *widgetPanel) rotate(time *canvas.Text) {
	c := software.NewTransparentCanvas()
	c.SetPadded(false)
	c.SetContent(time)

	img := c.Capture()
	out := imaging.Rotate270(img)

	w.rotated.Image = out
	ratio := time.MinSize().Width / time.MinSize().Height
	space := wmtheme.NarrowBarWidth - theme.Padding()*2
	w.rotated.SetMinSize(fyne.NewSize(space, space*ratio))
	w.rotated.Refresh()
}

func (w *widgetPanel) CreateRenderer() fyne.WidgetRenderer {
	narrow := w.desk.Settings().NarrowWidgetPanel()
	accountLabel := w.accountLabel()
	var account *widget.Button
	w.account = widget.NewButtonWithIcon(accountLabel, wmtheme.UserIcon, func() {
		w.showAccountMenu(account)
	})

	w.rotated = &canvas.Image{}
	w.clocks = container.NewMax(w.clock, container.New(&vClockPad{}, w.rotated))
	if narrow {
		w.clock.Hide()
	} else {
		w.clocks.Objects[1].Hide()
	}
	w.search = widget.NewButtonWithIcon("", theme.SearchIcon(), ShowAppLauncher)
	bottom := container.NewBorder(nil, nil, w.search, nil, w.account)

	bg := canvas.NewRectangle(wmtheme.WidgetPanelBackground())
	objects := []fyne.CanvasObject{
		bg,
		canvas.NewRectangle(color.Transparent), // clear top edge for clocks
		w.clocks,
		w.date,
		w.notifications}

	w.modules = container.NewVBox()
	objects = append(objects, layout.NewSpacer(), w.modules, bottom)
	w.loadModules(w.desk.Modules())

	return &widgetRenderer{
		panel:   w,
		bg:      bg,
		layout:  layout.NewVBoxLayout(),
		objects: objects,
	}
}

func (w *widgetPanel) MinSize() fyne.Size {
	if w.desk.Settings().NarrowWidgetPanel() {
		return fyne.NewSize(wmtheme.NarrowBarWidth, 200)
	}
	return fyne.NewSize(wmtheme.WidgetPanelWidth, 200)
}

func (w *widgetPanel) accountLabel() string {
	if w.desk.Settings().NarrowWidgetPanel() {
		return ""
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		fyne.LogError("Unable to look up user", err)
		return "Account"
	}

	return path.Base(homedir)
}

func (w *widgetPanel) reloadModules(mods []fynedesk.Module) {
	w.modules.Objects = nil
	w.loadModules(mods)
	w.modules.Refresh()
}

func (w *widgetPanel) loadModules(mods []fynedesk.Module) {
	for _, m := range mods {
		if statusMod, ok := m.(fynedesk.StatusAreaModule); ok {
			wid := statusMod.StatusAreaWidget()
			if wid == nil {
				continue
			}

			w.modules.Objects = append(w.modules.Objects, wid)
		}
	}
}

func newWidgetPanel(rootDesk fynedesk.Desktop) *widgetPanel {
	w := &widgetPanel{desk: rootDesk}
	w.ExtendBaseWidget(w)
	w.notifications = startNotifications()
	w.createClock()

	return w
}

type vClockPad struct {
}

func (u *vClockPad) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	objects[0].Resize(objects[0].MinSize())
	objects[0].Move(fyne.NewPos(5, 0))
}

func (u *vClockPad) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return objects[0].MinSize().Subtract(fyne.NewSize(0, theme.Padding()))
}
