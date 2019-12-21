package desktop

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	wmtheme "fyne.io/desktop/theme"
	"fyne.io/fyne"
	"fyne.io/fyne/cmd/fyne_settings/settings"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

// DeskSettings describes the configuration options available for Fyne desktop
type DeskSettings interface {
	Background() string
	IconTheme() string
	LauncherIcons() []string
	LauncherIconSize() int
	LauncherDisableTaskbar() bool
	LauncherDisableZoom() bool
	LauncherZoomScale() float64
	AddChangeListener(listener chan DeskSettings)
}

type deskSettings struct {
	background             string
	iconTheme              string
	launcherIcons          []string
	launcherIconSize       int
	launcherDisableTaskbar bool
	launcherDisableZoom    bool
	launcherZoomScale      float64

	listenerLock    sync.Mutex
	changeListeners []chan DeskSettings
}

const randrHelper = "arandr"

func (d *deskSettings) Background() string {
	return d.background
}

func (d *deskSettings) IconTheme() string {
	return d.iconTheme
}

func (d *deskSettings) LauncherIcons() []string {
	return d.launcherIcons
}

func (d *deskSettings) LauncherIconSize() int {
	return d.launcherIconSize
}

func (d *deskSettings) LauncherDisableTaskbar() bool {
	return d.launcherDisableTaskbar
}

func (d *deskSettings) LauncherDisableZoom() bool {
	return d.launcherDisableZoom
}

func (d *deskSettings) LauncherZoomScale() float64 {
	return d.launcherZoomScale
}

func (d *deskSettings) AddChangeListener(listener chan DeskSettings) {
	d.listenerLock.Lock()
	defer d.listenerLock.Unlock()
	d.changeListeners = append(d.changeListeners, listener)
}

func (d *deskSettings) apply() {
	d.listenerLock.Lock()
	defer d.listenerLock.Unlock()

	for _, listener := range d.changeListeners {
		select {
		case listener <- d:
		default:
			l := listener
			go func() { l <- d }()
		}
	}
}

func (d *deskSettings) setBackground(name string) {
	d.background = name
	fyne.CurrentApp().Preferences().SetString("background", d.background)
	d.apply()
}

func (d *deskSettings) setIconTheme(name string) {
	d.iconTheme = name
	fyne.CurrentApp().Preferences().SetString("icontheme", d.iconTheme)
	d.apply()
}

func (d *deskSettings) setLauncherIcons(defaultApps []string) {
	newLauncherIcons := strings.Join(defaultApps, "|")
	d.launcherIcons = defaultApps
	fyne.CurrentApp().Preferences().SetString("launchericons", newLauncherIcons)
	d.apply()
}

func (d *deskSettings) setLauncherIconSize(size int) {
	d.launcherIconSize = size
	fyne.CurrentApp().Preferences().SetInt("launchericonsize", d.launcherIconSize)
	d.apply()
}

func (d *deskSettings) setLauncherDisableTaskbar(taskbar bool) {
	d.launcherDisableTaskbar = taskbar
	fyne.CurrentApp().Preferences().SetBool("launcherdisabletaskbar", d.launcherDisableTaskbar)
	d.apply()
}

func (d *deskSettings) setLauncherDisableZoom(zoom bool) {
	d.launcherDisableZoom = zoom
	fyne.CurrentApp().Preferences().SetBool("launcherdisablezoom", d.launcherDisableZoom)
	d.apply()
}

func (d *deskSettings) setLauncherZoomScale(scale float64) {
	d.launcherZoomScale = scale
	fyne.CurrentApp().Preferences().SetFloat("launcherzoomscale", d.launcherZoomScale)
	d.apply()
}

func (d *deskSettings) load() {
	env := os.Getenv("FYNEDESK_BACKGROUND")
	if env != "" {
		d.background = env
	} else {
		d.background = fyne.CurrentApp().Preferences().String("background")
	}

	env = os.Getenv("FYNEDESK_ICONTHEME")
	if env != "" {
		d.iconTheme = env
	} else {
		d.iconTheme = fyne.CurrentApp().Preferences().String("icontheme")
	}
	if d.iconTheme == "" {
		d.iconTheme = "hicolor"
	}

	launcherIcons := fyne.CurrentApp().Preferences().String("launchericons")
	if launcherIcons != "" {
		d.launcherIcons = strings.SplitN(fyne.CurrentApp().Preferences().String("launchericons"), "|", -1)
	}
	if len(d.launcherIcons) == 0 {
		defaultApps := Instance().IconProvider().DefaultApps()
		for _, appData := range defaultApps {
			d.launcherIcons = append(d.launcherIcons, appData.Name())
		}
	}

	d.launcherIconSize = fyne.CurrentApp().Preferences().Int("launchericonsize")
	if d.launcherIconSize == 0 {
		d.launcherIconSize = 32
	}

	d.launcherDisableTaskbar = fyne.CurrentApp().Preferences().Bool("launcherdisabletaskbar")
	d.launcherDisableZoom = fyne.CurrentApp().Preferences().Bool("launcherdisablezoom")

	d.launcherZoomScale = fyne.CurrentApp().Preferences().Float("launcherzoomscale")
	if d.launcherZoomScale == 0.0 {
		d.launcherZoomScale = 2.0
	}
}

func (d *deskSettings) populateThemeIcons(box *fyne.Container, theme string) {
	box.Objects = nil
	for _, appName := range d.launcherIcons {
		appData := Instance().IconProvider().FindAppFromName(appName)
		iconRes := appData.Icon(theme, int((float64(d.launcherIconSize)*d.launcherZoomScale)*float64(Instance().Root().Canvas().Scale())))
		icon := widget.NewIcon(iconRes)
		box.AddObject(icon)
	}
	box.Refresh()
}

func (d *deskSettings) loadAppearanceScreen() fyne.CanvasObject {
	bgEntry := widget.NewEntry()
	if fyne.CurrentApp().Preferences().String("background") == "" {
		bgEntry.SetPlaceHolder("Input A File Path")
	} else {
		bgEntry.SetText(fyne.CurrentApp().Preferences().String("background"))
	}
	bgLabel := widget.NewLabelWithStyle("Background", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	themeLabel := widget.NewLabel(d.iconTheme)
	themeIcons := fyne.NewContainerWithLayout(layout.NewHBoxLayout())
	d.populateThemeIcons(themeIcons, d.iconTheme)
	themeList := fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	for _, themeName := range Instance().IconProvider().AvailableThemes() {
		themeButton := widget.NewButton(themeName, nil)
		themeButton.OnTapped = func() {
			themeLabel.SetText(themeButton.Text)
			d.populateThemeIcons(themeIcons, themeButton.Text)
		}
		themeList.AddObject(themeButton)
	}
	top := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, bgLabel, nil), bgLabel, bgEntry)

	themeFormLabel := widget.NewLabelWithStyle("Icon Theme", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	themeCurrent := widget.NewHBox(layout.NewSpacer(), themeLabel, themeIcons)
	middle := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, themeCurrent, themeFormLabel, nil),
		themeCurrent, themeFormLabel, widget.NewScrollContainer(themeList))

	applyButton := widget.NewHBox(layout.NewSpacer(),
		&widget.Button{Text: "Apply", Style: widget.PrimaryButton, OnTapped: func() {
			d.setBackground(bgEntry.Text)
			d.setIconTheme(themeLabel.Text)
		}})

	return fyne.NewContainerWithLayout(layout.NewBorderLayout(top, applyButton, nil, nil), top, applyButton, middle)
}

func (d *deskSettings) listAppMatches(lookupList *fyne.Container, orderList *fyne.Container, input string, launcherIcons []string) {
	desk := Instance()
	dataRange := desk.IconProvider().FindAppsMatching(input)
	for _, data := range dataRange {
		appData := data
		if appData.Name() == "" {
			continue
		}
		check := widget.NewCheck(appData.Name(), nil)
		iconRes := appData.Icon(d.iconTheme, int((float64(d.launcherIconSize)*d.launcherZoomScale)*float64(desk.Root().Canvas().Scale())))
		icon := widget.NewIcon(iconRes)
		hbox := widget.NewHBox(icon, check)
		exists := false
		for _, defaultApp := range launcherIcons {
			if defaultApp == appData.Name() {
				exists = true
				break
			}
		}
		if exists {
			check.SetChecked(true)
		} else {
			check.SetChecked(false)
		}
		check.OnChanged = func(checked bool) {
			if checked {
				launcherIcons = append(launcherIcons, check.Text)
			} else {
				index := -1
				for i, defaultApp := range launcherIcons {
					if defaultApp == check.Text {
						index = i
						break
					}
				}
				if index >= 0 {
					launcherIcons = append(launcherIcons[:index], launcherIcons[index+1:]...)
				}
			}
			d.populateOrderList(orderList, launcherIcons)
		}
		lookupList.AddObject(hbox)
	}
}

func (d *deskSettings) populateOrderList(list *fyne.Container, launcherIcons []string) {
	list.Objects = nil
	buttonMap := make(map[fyne.CanvasObject]int)
	for i, appName := range launcherIcons {
		appData := Instance().IconProvider().FindAppFromName(appName)
		upButton := widget.NewButtonWithIcon("", theme.MoveUpIcon(), nil)
		buttonMap[upButton] = i
		upButton.OnTapped = func() {
			index := buttonMap[upButton]
			if index > 0 {
				launcherIcons[index-1], launcherIcons[index] = launcherIcons[index], launcherIcons[index-1]
				d.populateOrderList(list, launcherIcons)
			}
		}
		downButton := widget.NewButtonWithIcon("", theme.MoveDownIcon(), nil)
		buttonMap[downButton] = i
		downButton.OnTapped = func() {
			index := buttonMap[downButton]
			if index < len(d.launcherIcons)-1 {
				launcherIcons[index+1], launcherIcons[index] = launcherIcons[index], launcherIcons[index+1]
				d.populateOrderList(list, launcherIcons)
			}
		}
		iconRes := appData.Icon(d.iconTheme, int((float64(d.launcherIconSize)*d.launcherZoomScale)*float64(Instance().Root().Canvas().Scale())))
		icon := widget.NewIcon(iconRes)
		label := widget.NewLabel(appName)
		hbox := widget.NewHBox(upButton, downButton, icon, label)
		list.AddObject(hbox)
	}
	list.Refresh()
}

func (d *deskSettings) loadBarScreen() fyne.CanvasObject {
	var launcherIcons []string
	launcherIcons = append(launcherIcons, d.launcherIcons...)

	lookupList := fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	orderList := fyne.NewContainerWithLayout(layout.NewVBoxLayout())

	entry := widget.NewEntry()
	entry.SetPlaceHolder("Start Typing An Application Name")
	entry.OnChanged = func(input string) {
		lookupList.Objects = nil
		if input == "" {
			return
		}

		d.listAppMatches(lookupList, orderList, input, launcherIcons)
	}
	d.populateOrderList(orderList, launcherIcons)

	lookup := fyne.NewContainerWithLayout(layout.NewBorderLayout(entry, nil, nil, nil), entry, widget.NewScrollContainer(lookupList))
	order := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, nil, nil), widget.NewScrollContainer(orderList))

	iconSize := widget.NewEntry()
	iconSize.SetText(fmt.Sprintf("%d", d.launcherIconSize))

	zoomScale := widget.NewEntry()
	zoomScale.SetText(fmt.Sprintf("%f", d.launcherZoomScale))

	top := widget.NewForm(
		&widget.FormItem{Text: "Launcher Icon Size:", Widget: iconSize},
		&widget.FormItem{Text: "Launcher Zoom Scale:", Widget: zoomScale})

	disableTaskbar := widget.NewCheck("Disable Taskbar", nil)
	disableTaskbar.SetChecked(d.launcherDisableTaskbar)

	disableZoom := widget.NewCheck("Disable Zoom", nil)
	disableZoom.SetChecked(d.launcherDisableZoom)

	settings := widget.NewVBox(top, disableTaskbar, disableZoom)

	tabs := widget.NewTabContainer(widget.NewTabItem("Launcher Icons", lookup), widget.NewTabItem("Launcher Icon Order", order),
		widget.NewTabItem("Launcher Settings", settings))
	applyButton := widget.NewHBox(layout.NewSpacer(),
		&widget.Button{Text: "Apply", Style: widget.PrimaryButton, OnTapped: func() {
			size, err := strconv.Atoi(iconSize.Text)
			if err != nil {
				fyne.LogError("error setting launcher icon size", err)
				size = 32
			}
			d.setLauncherIconSize(size)

			scale, err := strconv.ParseFloat(zoomScale.Text, 32)
			if err != nil {
				fyne.LogError("Error setting launcher zoom scale", err)
				scale = 2.0
			}
			d.setLauncherZoomScale(scale)
			d.setLauncherDisableTaskbar(disableTaskbar.Checked)
			d.setLauncherDisableZoom(disableZoom.Checked)

			d.setLauncherIcons(launcherIcons)
		}})

	barSettings := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, applyButton, nil, nil), applyButton, tabs)

	return barSettings
}

func loadAdvancedScreen() fyne.CanvasObject {
	var displays fyne.CanvasObject
	if _, err := exec.LookPath(randrHelper); err == nil {
		displays = widget.NewButtonWithIcon("Displays", wmtheme.DisplayIcon, func() {
			exec.Command(randrHelper).Start()
		})
	} else {
		displays = widget.NewLabel("This requires " + randrHelper)
	}

	return widget.NewVBox(displays)
}

func showSettings(deskSettings *deskSettings) {
	w := fyne.CurrentApp().NewWindow("FyneDesk Settings")
	fyneSettings := settings.NewSettings()

	tabs := widget.NewTabContainer(
		&widget.TabItem{Text: "Fyne Settings", Icon: theme.FyneLogo(),
			Content: fyneSettings.LoadAppearanceScreen()},
		&widget.TabItem{Text: "Appearance", Icon: fyneSettings.AppearanceIcon(),
			Content: deskSettings.loadAppearanceScreen()},
		&widget.TabItem{Text: "App Bar", Icon: wmtheme.IconifyIcon, Content: deskSettings.loadBarScreen()},
		&widget.TabItem{Text: "Advanced", Icon: theme.SettingsIcon(),
			Content: loadAdvancedScreen()},
	)
	tabs.SetTabLocation(widget.TabLocationLeading)
	w.SetContent(tabs)

	w.Resize(fyne.NewSize(480, 320))
	w.Show()
}

// NewDeskSettings loads the user's preferences from environment or config
func NewDeskSettings() DeskSettings {
	settings := &deskSettings{}
	settings.load()

	return settings
}
