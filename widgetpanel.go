package desktop

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	wmtheme "github.com/fyne-io/desktop/theme"
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

	clock         *canvas.Text
	date, battery *widget.Label
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

func (w *widgetPanel) batteryTick() {
	tick := time.NewTicker(time.Second * 10)
	go func() {
		for {
			w.battery.SetValue(battery())
			<-tick.C
		}
	}()
}

func formattedTime() string {
	return time.Now().Format("15:04pm")
}

func formattedDate() string {
	return time.Now().Format("2 January")
}

func battery() float64 {
	nowStr, err1 := ioutil.ReadFile("/sys/class/power_supply/BAT0/charge_now")
	fullStr, err2 := ioutil.ReadFile("/sys/class/power_supply/BAT0/charge_full")
	if err1 != nil || err2 != nil {
		log.Println("Error reading battery info", err1)
		return 0
	}

	now, err1 := strconv.Atoi(strings.TrimSpace(string(nowStr)))
	full, err2 := strconv.Atoi(strings.TrimSpace(string(fullStr)))
	if err1 != nil || err2 != nil {
		log.Println("Error converting battery info", err1)
		return 0
	}

	return float64(now)/float64(full)
}

func (w *widgetPanel) createBattery() {
	w.battery = widget.NewProgressBar()
	w.brightness = widget.NewProgressBar()

	go w.batteryTick()
}

func (w *widgetPanel) createClock() {
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
}

func (w *widgetPanel) CreateRenderer() fyne.WidgetRenderer {
	themes := fyne.NewContainerWithLayout(layout.NewGridLayout(2),
		widget.NewButton("Light", func() {
			fyne.CurrentApp().Settings().SetTheme(theme.LightTheme())
			_ = os.Setenv("FYNE_THEME", "light")
		}),
		widget.NewButton("Dark", func() {
			fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme())
			_ = os.Setenv("FYNE_THEME", "dark")
		}))
	reload := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		os.Exit(1)
	})

	quit := widget.NewButton("Log Out", func() {
		w.root.Close()
	})
	buttons := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, reload, nil),
		reload, quit)

	batteryIcon := widget.NewIcon(wmtheme.BatteryIcon)

	objects := []fyne.CanvasObject{
		w.clock,
		w.date,
		layout.NewSpacer(),
		fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, batteryIcon, nil), batteryIcon, w.battery),
		themes,
		buttons,
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
	w.createBattery()

	return w
}
