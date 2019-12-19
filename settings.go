package desktop

import (
	"os"
	"os/exec"
	"strings"

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
	DefaultApps() []string
}

type deskSettings struct {
	background  string
	iconTheme   string
	defaultApps []string
}

const randrHelper = "arandr"

func (d *deskSettings) Background() string {
	return d.background
}

func (d *deskSettings) IconTheme() string {
	return d.iconTheme
}

func (d *deskSettings) DefaultApps() []string {
	return d.defaultApps
}

func (d *deskSettings) setBackground(name string) {
	d.background = name
	fyne.CurrentApp().Preferences().SetString("background", d.background)
	go Instance().(*deskLayout).updateBackgrounds()
}

func (d *deskSettings) setIconTheme(name string) {
	d.iconTheme = name
	fyne.CurrentApp().Preferences().SetString("icontheme", d.iconTheme)
	go Instance().(*deskLayout).updateIconTheme()
}

func (d *deskSettings) setDefaultApps(defaultApps []string) {
	newDefaultApps := strings.Join(defaultApps, "|")
	d.defaultApps = defaultApps
	fyne.CurrentApp().Preferences().SetString("defaultapps", newDefaultApps)
	go Instance().(*deskLayout).updateIconOrder()
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

	d.defaultApps = strings.SplitN(fyne.CurrentApp().Preferences().String("defaultapps"), "|", -1)

	if len(d.defaultApps) == 0 {
		defaultApps := Instance().IconProvider().DefaultApps()
		for _, appData := range defaultApps {
			d.defaultApps = append(d.defaultApps, appData.Name())
		}
	}
}

func (d *deskSettings) populateThemeIcons(box *fyne.Container) {
	box.Objects = nil
	for _, appName := range d.defaultApps {
		appData := Instance().IconProvider().FindAppFromName(appName)
		iconRes := appData.Icon(d.iconTheme, int(32.0*fyne.CurrentApp().Settings().Scale()))
		icon := widget.NewIcon(iconRes)
		box.AddObject(icon)
	}
	box.Refresh()
}

func (d *deskSettings) loadAppearanceScreen() fyne.CanvasObject {
	bgEntry := widget.NewEntry()
	if fyne.CurrentApp().Preferences().String("background") == "" {
		bgEntry.SetText("Default")
	} else {
		bgEntry.SetText(fyne.CurrentApp().Preferences().String("background"))
	}
	themeLabel := widget.NewLabel(d.iconTheme + ":")
	themeIcons := fyne.NewContainerWithLayout(layout.NewHBoxLayout())
	d.populateThemeIcons(themeIcons)

	themeList := fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	for _, themeName := range Instance().IconProvider().AvailableThemes() {
		themeButton := widget.NewButton(themeName, nil)
		themeButton.OnTapped = func() {
			themeLabel.SetText(themeButton.Text + ":")
			d.setIconTheme(themeButton.Text)
			d.populateThemeIcons(themeIcons)
		}
		themeList.AddObject(themeButton)
	}
	top := widget.NewVBox(widget.NewHBox(widget.NewLabelWithStyle("Background", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}), bgEntry),
		widget.NewHBox(layout.NewSpacer(),
			&widget.Button{Text: "Apply", Style: widget.PrimaryButton, OnTapped: func() {
				d.setBackground(bgEntry.Text)
			}}))

	themeFormLabel := widget.NewLabelWithStyle("Icon Theme", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	themeCurrent := widget.NewHBox(themeLabel, themeIcons)
	middle := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, themeCurrent, themeFormLabel, nil),
		themeCurrent, themeFormLabel, widget.NewScrollContainer(themeList))

	return fyne.NewContainerWithLayout(layout.NewBorderLayout(top, nil, nil, nil), top, middle)
}

func (d *deskSettings) listAppMatches(lookupList *fyne.Container, orderList *fyne.Container, input string) {
	desk := Instance()
	dataRange := desk.IconProvider().FindAppsMatching(input)
	defaultApps := desk.Settings().DefaultApps()
	for _, data := range dataRange {
		appData := data
		if appData.Name() == "" {
			continue
		}
		check := widget.NewCheck(appData.Name(), nil)
		iconRes := appData.Icon(d.iconTheme, int(32.0*fyne.CurrentApp().Settings().Scale()))
		icon := widget.NewIcon(iconRes)
		hbox := widget.NewHBox(icon, check)
		exists := false
		for _, defaultApp := range defaultApps {
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
				defaultApps = append(defaultApps, check.Text)
			} else {
				index := -1
				for i, defaultApp := range defaultApps {
					if defaultApp == check.Text {
						index = i
						break
					}
				}
				if index >= 0 {
					defaultApps = append(defaultApps[:index], defaultApps[index+1:]...)
				}
			}
			d.setDefaultApps(defaultApps)
			d.populateOrderList(orderList)
		}
		lookupList.AddObject(hbox)
	}
}

func (d *deskSettings) populateOrderList(list *fyne.Container) {
	list.Objects = nil
	buttonMap := make(map[fyne.CanvasObject]int)
	for i, appName := range d.defaultApps {
		appData := Instance().IconProvider().FindAppFromName(appName)
		upButton := widget.NewButtonWithIcon("", theme.MoveUpIcon(), nil)
		buttonMap[upButton] = i
		upButton.OnTapped = func() {
			index := buttonMap[upButton]
			if index > 0 {
				d.defaultApps[index-1], d.defaultApps[index] = d.defaultApps[index], d.defaultApps[index-1]
				d.setDefaultApps(d.defaultApps)
				d.populateOrderList(list)
			}
		}
		downButton := widget.NewButtonWithIcon("", theme.MoveDownIcon(), nil)
		buttonMap[downButton] = i
		downButton.OnTapped = func() {
			index := buttonMap[downButton]
			if index < len(d.defaultApps)-1 {
				d.defaultApps[index+1], d.defaultApps[index] = d.defaultApps[index], d.defaultApps[index+1]
				d.setDefaultApps(d.defaultApps)
				d.populateOrderList(list)
			}
		}
		iconRes := appData.Icon(d.iconTheme, int(32.0*fyne.CurrentApp().Settings().Scale()))
		icon := widget.NewIcon(iconRes)
		label := widget.NewLabel(appName)
		hbox := widget.NewHBox(upButton, downButton, icon, label)
		list.AddObject(hbox)
	}
	list.Refresh()
}

func (d *deskSettings) loadBarScreen() fyne.CanvasObject {
	lookupList := fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	orderList := fyne.NewContainerWithLayout(layout.NewVBoxLayout())

	entry := widget.NewEntry()
	entry.SetPlaceHolder("Start Typing An Application Name")
	entry.OnChanged = func(input string) {
		lookupList.Objects = nil
		if input == "" {
			return
		}

		d.listAppMatches(lookupList, orderList, input)
	}

	d.populateOrderList(orderList)

	lookup := fyne.NewContainerWithLayout(layout.NewBorderLayout(entry, nil, nil, nil), entry, widget.NewScrollContainer(lookupList))
	order := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, nil, nil), widget.NewScrollContainer(orderList))
	barSettings := widget.NewTabContainer(widget.NewTabItem("Default Icons", lookup), widget.NewTabItem("Icon Order", order))

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
