package desktop

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	wmtheme "fyne.io/desktop/theme"
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
	baseWidget

	desk *deskLayout
	root fyne.Window

	clock               *canvas.Text
	date                *widget.Label
	battery, brightness *widget.ProgressBar
}

func (w *widgetPanel) Hide() {
	w.Hidden = true

	canvas.Refresh(w)
}

func (w *widgetPanel) MinSize() fyne.Size {
	return widget.Renderer(w).MinSize()
}

func (w *widgetPanel) Move(pos fyne.Position) {
	w.position = pos

	canvas.Refresh(w)
}

func (w *widgetPanel) Position() fyne.Position {
	return w.position
}

func (w *widgetPanel) Resize(size fyne.Size) {
	w.size = size

	widget.Renderer(w).Layout(size)
}

func (w *widgetPanel) Show() {
	w.Hidden = false

	canvas.Refresh(w)
}

func appExecPopUpListMatches(w *widgetPanel, popup *widget.PopUp, appList *fyne.Container, input string) {
	iconTheme := w.desk.Settings().IconTheme()
	dataRange := w.desk.IconProvider().FindIconsMatchingAppName(iconTheme, iconSize, input)
	if len(dataRange) == 0 {
		return
	}
	for _, data := range dataRange {
		if data == nil || data.IconPath() == "" {
			continue
		}
		bytes, err := ioutil.ReadFile(data.IconPath())
		if err != nil {
			fyne.LogError("Could not read file", err)
			continue
		}
		str := strings.Replace(data.IconPath(), "-", "", -1)
		iconResource := strings.Replace(str, "_", "", -1)

		res := fyne.NewStaticResource(strings.ToLower(filepath.Base(iconResource)), bytes)
		app := widget.NewButtonWithIcon(data.Name(), res, func() {
			command := strings.Split(data.Exec(), " ")
			exec.Command(command[0]).Start()
			popup.Hide()
			popup = nil
		})
		appList.AddObject(app)
	}
}

func appExecPopUp(w *widgetPanel) {
	entry := widget.NewEntry()
	entry.SetPlaceHolder("Application")

	appList := fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	appScroller := widget.NewScrollContainer(appList)

	sizingRect := canvas.NewRectangle(color.Transparent)
	sizingRect.SetMinSize(fyne.NewSize(int(w.root.Canvas().Size().Width/4), iconSize*3))

	sizer := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, nil, nil), sizingRect, appScroller)
	content := fyne.NewContainerWithLayout(layout.NewVBoxLayout(), entry, sizer)
	popup := widget.NewModalPopUp(content, w.root.Canvas())

	cancel := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		popup.Hide()
		popup = nil
	})
	content.AddObject(cancel)

	entry.OnChanged = func(input string) {
		appList.Objects = nil
		if input != "" {
			appExecPopUpListMatches(w, popup, appList, input)
		}
		canvas.Refresh(appList)
	}

	popup.Show()
	w.root.RequestFocus()
	w.root.Canvas().Focus(entry)
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
			w.brightness.SetValue(brightness())
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

	return float64(now) / float64(full)
}

func (w *widgetPanel) createBattery() {
	w.battery = widget.NewProgressBar()
	w.brightness = widget.NewProgressBar()

	go w.batteryTick()
}

func brightness() float64 {
	out, err := exec.Command("xbacklight").Output()
	if err != nil {
		log.Println("Error running xbacklight", err)
	} else {
		ret, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
		if err != nil {
			log.Println("Error reading brightness info", err)
			return 0
		}
		return float64(ret) / 100
	}
	return 0
}

func (w *widgetPanel) setBrightness(diff int) {
	value := int(brightness()*100) + diff

	if value < 5 {
		value = 5
	} else if value > 100 {
		value = 100
	}

	err := exec.Command("xbacklight", "-set", fmt.Sprintf("%d", value)).Run()
	if err != nil {
		log.Println("Error running xbacklight", err)
	} else {
		w.brightness.SetValue(brightness())
	}
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
	brightnessIcon := widget.NewIcon(wmtheme.BrightnessIcon)
	less := widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() {
		w.setBrightness(-5)
	})
	more := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		w.setBrightness(5)
	})
	bright := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, less, more),
		less, w.brightness, more)
	appExecButton := widget.NewButtonWithIcon("Applications", theme.SearchIcon(), func() {
		appExecPopUp(w)
	})
	objects := []fyne.CanvasObject{
		w.clock,
		w.date,
		layout.NewSpacer(),
		appExecButton,
		fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, batteryIcon, nil), batteryIcon, w.battery),
		fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, brightnessIcon, nil), brightnessIcon, bright),
		themes,
		buttons,
	}

	return &widgetRenderer{
		panel:   w,
		layout:  layout.NewVBoxLayout(),
		objects: objects,
	}
}

func newWidgetPanel(rootDesk *deskLayout) *widgetPanel {
	w := &widgetPanel{
		desk: rootDesk,
		root: rootDesk.win,
	}
	w.createClock()
	w.createBattery()

	return w
}
