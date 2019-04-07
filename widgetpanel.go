package desktop

import (
	"image/color"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

type widgetRenderer struct {
	panel *widgetPanel

	layout  fyne.Layout
	objects []fyne.CanvasObject
}

func (w *widgetRenderer) MinSize() fyne.Size {
	return w.layout.MinSize(w.objects)
}

func (w *widgetRenderer) Layout(size fyne.Size) {
	w.layout.Layout(w.objects, size)
}

func (w *widgetRenderer) Refresh() {
}

func (w *widgetRenderer) ApplyTheme() {
	w.panel.clock.Color = theme.TextColor()
	canvas.Refresh(w.panel.clock)
}

func (w *widgetRenderer) BackgroundColor() color.Color {
	r, g, b, _ := theme.BackgroundColor().RGBA()
	return &color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 0x99}
}

func (w *widgetRenderer) Objects() []fyne.CanvasObject {
	return w.objects
}

func (w *widgetRenderer) Destroy() {
}

type widgetPanel struct {
	root   fyne.Window
	size   fyne.Size
	pos    fyne.Position
	hidden bool

	clock *canvas.Text
	date  *widget.Label
}

func (w *widgetPanel) Hide() {
	w.hidden = true

	canvas.Refresh(w)
}

func (w *widgetPanel) MinSize() fyne.Size {
	return widget.Renderer(w).MinSize()
}

func (w *widgetPanel) Move(pos fyne.Position) {
	w.pos = pos

	canvas.Refresh(w)
}

func (w *widgetPanel) Position() fyne.Position {
	return w.pos
}

func (w *widgetPanel) Resize(size fyne.Size) {
	w.size = size

	widget.Renderer(w).Layout(size)
}

func (w *widgetPanel) Show() {
	w.hidden = false

	canvas.Refresh(w)
}

func (w *widgetPanel) Size() fyne.Size {
	return w.size
}

func (w *widgetPanel) Visible() bool {
	return !w.hidden
}

func (w *widgetPanel) clockTick() {
	tick := time.NewTicker(time.Second)
	go func() {
		for {
			<-tick.C
			w.clock.Text = formattedTime()
			canvas.Refresh(w.clock)

			w.date.SetText(formattedDate())
			canvas.Refresh(w.date)
		}
	}()
}

func formattedTime() string {
	return time.Now().Format("15:04pm")
}

func formattedDate() string {
	return time.Now().Format("2 January")
}

func (w *widgetPanel) createClock() *widget.Box {
	var style fyne.TextStyle
	style.Monospace = true

	w.clock = &canvas.Text{
		Color:     theme.TextColor(),
		Text:      formattedTime(),
		Alignment: fyne.TextAlignCenter,
		TextStyle: style,
		TextSize:  3 * theme.TextSize(),
	}
	w.date = &widget.Label{
		Text:      formattedDate(),
		Alignment: fyne.TextAlignCenter,
		TextStyle: style,
	}

	go w.clockTick()

	return widget.NewVBox(w.clock, w.date)
}

func (w *widgetPanel) CreateRenderer() fyne.WidgetRenderer {
	themes := fyne.NewContainerWithLayout(layout.NewGridLayout(2),
		widget.NewButton("Light", func() {
			fyne.CurrentApp().Settings().SetTheme(theme.LightTheme())
		}),
		widget.NewButton("Dark", func() {
			fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme())
		}))
	quit := widget.NewButton("Log Out", func() {
		w.root.Close()
	})

	objects := []fyne.CanvasObject{
		w.clock,
		w.date,
		layout.NewSpacer(),
		themes,
		quit,
	}

	return &widgetRenderer{
		panel:   w,
		layout:  layout.NewVBoxLayout(),
		objects: objects,
	}
}

func newWidgetPanel(rootWin fyne.Window) *widgetPanel {
	w := &widgetPanel{
		root: rootWin,
	}
	w.createClock()

	return w
}
