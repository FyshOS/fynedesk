package ui

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	"fyne.io/fynedesk"
	wmtheme "fyne.io/fynedesk/theme"
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
	widget.BaseWidget

	desk       fynedesk.Desktop
	root       fyne.Window
	appExecWin fyne.Window

	clock               *canvas.Text
	date                *widget.Label
	battery, brightness *widget.ProgressBar
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
			value, _ := battery()
			w.battery.SetValue(value)
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

func battery() (float64, error) {
	nowStr, err1 := ioutil.ReadFile("/sys/class/power_supply/BAT0/charge_now")
	fullStr, err2 := ioutil.ReadFile("/sys/class/power_supply/BAT0/charge_full")
	if err1 != nil || err2 != nil {
		log.Println("Error reading battery info", err1)
		return 0, err1
	}

	now, err1 := strconv.Atoi(strings.TrimSpace(string(nowStr)))
	full, err2 := strconv.Atoi(strings.TrimSpace(string(fullStr)))
	if err1 != nil || err2 != nil {
		log.Println("Error converting battery info", err1)
		return 0, err1
	}

	return float64(now) / float64(full), nil
}

func (w *widgetPanel) createBattery() {
	if _, err := battery(); err == nil {
		w.battery = widget.NewProgressBar()
		go w.batteryTick()
	}
	if _, err := brightness(); err == nil {
		w.brightness = widget.NewProgressBar()
	}
}

func brightness() (float64, error) {
	out, err := exec.Command("xbacklight").Output()
	if err != nil {
		log.Println("Error running xbacklight", err)
		return 0, err
	}
	ret, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		log.Println("Error reading brightness info", err)
		return 0, err
	}
	return float64(ret) / 100, nil
}

func (w *widgetPanel) setBrightness(diff int) {
	floatVal, _ := brightness()
	value := int(floatVal*100) + diff

	if value < 5 {
		value = 5
	} else if value > 100 {
		value = 100
	}

	err := exec.Command("xbacklight", "-set", fmt.Sprintf("%d", value)).Run()
	if err != nil {
		log.Println("Error running xbacklight", err)
	} else {
		newVal, _ := brightness()
		w.brightness.SetValue(newVal)
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

func (w *widgetPanel) showAccountMenu(from fyne.CanvasObject) {
	items := []*fyne.MenuItem{
		fyne.NewMenuItem("About", func() {
			showAbout()
		}),
		fyne.NewMenuItem("Settings", func() {
			showSettings(w.desk.Settings().(*deskSettings))
		}),
	}
	if w.desk.WindowManager() != nil {
		items = append(items, fyne.NewMenuItem("Blank Screen", w.desk.WindowManager().Blank))
	}
	if os.Getenv("FYNE_DESK_RUNNER") != "" && w.desk.(*deskLayout).wm != nil {
		items = append(items, fyne.NewMenuItem("Reload", func() {
			os.Exit(1)
		}))
	}

	closeLabel := "Log Out"
	if w.desk.(*deskLayout).wm == nil {
		closeLabel = "Quit"
	}
	items = append(items, fyne.NewMenuItem(closeLabel, func() {
		w.root.Close()
	}))

	popup := widget.NewPopUpMenu(fyne.NewMenu("Account", items...), w.root.Canvas())

	bottomLeft := fyne.CurrentApp().Driver().AbsolutePositionForObject(from)
	popup.Move(bottomLeft.Subtract(fyne.NewPos(0, popup.MinSize().Height)))
	popup.Resize(fyne.NewSize(from.Size().Width, popup.Content.MinSize().Height))
}

func (w *widgetPanel) CreateRenderer() fyne.WidgetRenderer {
	accountLabel := "Account"
	homedir, err := os.UserHomeDir()
	if err == nil {
		accountLabel = path.Base(homedir)
	} else {
		fyne.LogError("Unable to look up user", err)
	}
	var account *widget.Button
	account = widget.NewButtonWithIcon(accountLabel, wmtheme.UserIcon, func() {
		w.showAccountMenu(account)
	})
	appExecButton := widget.NewButtonWithIcon("Applications", theme.SearchIcon(), ShowAppLauncher)
	objects := []fyne.CanvasObject{
		w.clock,
		w.date,
		layout.NewSpacer(),
		appExecButton,
	}
	if _, err := battery(); err == nil {
		batteryIcon := widget.NewIcon(wmtheme.BatteryIcon)
		objects = append(objects,
			fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, batteryIcon, nil), batteryIcon, w.battery))
	}
	if _, err := brightness(); err == nil {
		brightnessIcon := widget.NewIcon(wmtheme.BrightnessIcon)
		less := widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() {
			w.setBrightness(-5)
		})
		more := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
			w.setBrightness(5)
		})
		bright := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, less, more),
			less, w.brightness, more)
		objects = append(objects,
			fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, brightnessIcon, nil), brightnessIcon, bright))
		go w.setBrightness(0)
	}
	objects = append(objects,
		account)

	return &widgetRenderer{
		panel:   w,
		layout:  layout.NewVBoxLayout(),
		objects: objects,
	}
}

func newWidgetPanel(rootDesk fynedesk.Desktop) *widgetPanel {
	w := &widgetPanel{
		desk:       rootDesk,
		root:       rootDesk.Root(),
		appExecWin: nil,
	}
	w.ExtendBaseWidget(w)
	w.createClock()
	w.createBattery()

	return w
}
