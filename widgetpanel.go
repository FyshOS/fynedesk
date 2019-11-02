package desktop

import (
	"fmt"
	"fyne.io/desktop/internal/settings"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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
	r, _, _, _ := theme.BackgroundColor().RGBA()
	if uint8(r) > 0x99 {
		return wmtheme.WidgetPanelBackgroundLight
	}
	return wmtheme.WidgetPanelBackgroundDark
}

func (w *widgetRenderer) Objects() []fyne.CanvasObject {
	return w.objects
}

func (w *widgetRenderer) Destroy() {
}

type widgetPanel struct {
	baseWidget

	desk       *deskLayout
	root       fyne.Window
	appExecWin fyne.Window

	clock               *canvas.Text
	date                *widget.Label
	battery, brightness *widget.ProgressBar
}

func (w *widgetPanel) Hide() {
	w.hide(w)
}

func (w *widgetPanel) MinSize() fyne.Size {
	return widget.Renderer(w).MinSize()
}

func (w *widgetPanel) Move(pos fyne.Position) {
	w.move(pos, w)
}

func (w *widgetPanel) Position() fyne.Position {
	return w.position
}

func (w *widgetPanel) Resize(size fyne.Size) {
	w.resize(size, w)
}

func (w *widgetPanel) Show() {
	w.show(w)
}

func appExecPopUpListMatches(w *widgetPanel, win fyne.Window, appList *fyne.Container, input string) {
	iconTheme := w.desk.Settings().IconTheme()
	dataRange := w.desk.IconProvider().FindAppsMatching(input)
	for _, data := range dataRange {
		appData := data                     // capture for goroutine below
		icon := appData.Icon(iconTheme, 32) // TODO match theme but FDO needs power of 2 theme.IconInlineSize())
		app := widget.NewButtonWithIcon(appData.Name(), icon, func() {
			err := w.desk.RunApp(appData)
			if err != nil {
				fyne.LogError("Failed to start app", err)
				return
			}
			win.Close()
		})
		appList.AddObject(app)
	}
}

func appExecPopUp(w *widgetPanel) fyne.Window {
	win := w.desk.app.NewWindow("Applications")
	appList := fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	appScroller := widget.NewScrollContainer(appList)

	entry := widget.NewEntry()
	entry.SetPlaceHolder("Application")
	entry.OnChanged = func(input string) {
		appList.Objects = nil
		if input == "" {
			return
		}

		appExecPopUpListMatches(w, win, appList, input)
	}

	cancel := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		win.Close()
	})

	content := fyne.NewContainerWithLayout(layout.NewBorderLayout(entry, cancel, nil, nil), entry, appScroller, cancel)

	win.SetContent(content)
	win.Resize(fyne.NewSize(300,
		cancel.MinSize().Height*4+theme.Padding()*6+entry.MinSize().Height))
	win.CenterOnScreen()
	win.Canvas().Focus(entry)
	return win
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

	settings := widget.NewButtonWithIcon("", theme.SettingsIcon(), settings.Show)
	reload := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		os.Exit(1)
	})
	if os.Getenv("FYNE_DESK_RUNNER") == "" {
		reload.Disable()
	}
	quit := widget.NewButton("Log Out", func() {
		w.root.Close()
	})
	buttons := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, reload, settings),
		reload, settings, quit)

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
		if w.appExecWin != nil {
			w.appExecWin.Close()
		}

		w.appExecWin = appExecPopUp(w)
		w.appExecWin.SetOnClosed(func() {
			w.appExecWin = nil
		})
		w.appExecWin.Show()
	})
	objects := []fyne.CanvasObject{
		w.clock,
		w.date,
		layout.NewSpacer(),
		appExecButton,
		fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, batteryIcon, nil), batteryIcon, w.battery),
	}

	if _, err := exec.LookPath("xbacklight"); err == nil {
		objects = append(objects,
			fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, brightnessIcon, nil), brightnessIcon, bright))
		go w.setBrightness(0)
	}
	objects = append(objects,
		themes,
		buttons)

	return &widgetRenderer{
		panel:   w,
		layout:  layout.NewVBoxLayout(),
		objects: objects,
	}
}

func newWidgetPanel(rootDesk *deskLayout) *widgetPanel {
	w := &widgetPanel{
		desk:       rootDesk,
		root:       rootDesk.win,
		appExecWin: nil,
	}
	w.createClock()
	w.createBattery()

	return w
}
