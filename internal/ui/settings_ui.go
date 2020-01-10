package ui

import (
	"fmt"
	"os/exec"
	"strconv"

	"fyne.io/desktop"
	wmtheme "fyne.io/desktop/theme"

	"fyne.io/fyne"
	"fyne.io/fyne/cmd/fyne_settings/settings"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

const randrHelper = "arandr"

type settingsUI struct {
	settings *deskSettings

	launcherIcons []string
}

func (d *settingsUI) populateThemeIcons(box *fyne.Container, theme string) {
	box.Objects = nil
	for _, appName := range d.launcherIcons {
		appData := desktop.Instance().IconProvider().FindAppFromName(appName)
		iconRes := appData.Icon(theme, int((float64(d.settings.LauncherIconSize())*d.settings.LauncherZoomScale())*float64(desktop.Instance().Screens().Primary().CanvasScale())))
		icon := widget.NewIcon(iconRes)
		box.AddObject(icon)
	}
	box.Refresh()
}

func (d *settingsUI) loadAppearanceScreen() fyne.CanvasObject {
	bgEntry := widget.NewEntry()
	if fyne.CurrentApp().Preferences().String("background") == "" {
		bgEntry.SetPlaceHolder("Input A File Path")
	} else {
		bgEntry.SetText(fyne.CurrentApp().Preferences().String("background"))
	}
	bgLabel := widget.NewLabelWithStyle("Background", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	themeLabel := widget.NewLabel(d.settings.IconTheme())
	themeIcons := fyne.NewContainerWithLayout(layout.NewHBoxLayout())
	d.populateThemeIcons(themeIcons, d.settings.IconTheme())
	themeList := fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	for _, themeName := range desktop.Instance().IconProvider().AvailableThemes() {
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
			d.settings.setBackground(bgEntry.Text)
			d.settings.setIconTheme(themeLabel.Text)
		}})

	return fyne.NewContainerWithLayout(layout.NewBorderLayout(top, applyButton, nil, nil), top, applyButton, middle)
}

func (d *settingsUI) listAppMatches(lookupList *fyne.Container, orderList *fyne.Container, input string) {
	desk := desktop.Instance()
	dataRange := desk.IconProvider().FindAppsMatching(input)
	for _, data := range dataRange {
		appData := data
		if appData.Name() == "" {
			continue
		}
		check := widget.NewCheck(appData.Name(), nil)
		iconRes := appData.Icon(d.settings.IconTheme(), int((float64(d.settings.LauncherIconSize())*d.settings.LauncherZoomScale())*float64(desk.Screens().Primary().CanvasScale())))
		icon := widget.NewIcon(iconRes)
		hbox := widget.NewHBox(icon, check)
		exists := false
		for _, defaultApp := range d.launcherIcons {
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
				d.launcherIcons = append(d.launcherIcons, check.Text)
			} else {
				index := -1
				for i, defaultApp := range d.launcherIcons {
					if defaultApp == check.Text {
						index = i
						break
					}
				}
				if index >= 0 {
					d.launcherIcons = append(d.launcherIcons[:index], d.launcherIcons[index+1:]...)
				}
			}
			d.populateOrderList(orderList)
		}
		lookupList.AddObject(hbox)
	}
}

func (d *settingsUI) populateOrderList(list *fyne.Container) {
	list.Objects = nil
	buttonMap := make(map[fyne.CanvasObject]int)
	for i, appName := range d.launcherIcons {
		appData := desktop.Instance().IconProvider().FindAppFromName(appName)
		upButton := widget.NewButtonWithIcon("", theme.MoveUpIcon(), nil)
		buttonMap[upButton] = i
		upButton.OnTapped = func() {
			index := buttonMap[upButton]
			if index > 0 {
				d.launcherIcons[index-1], d.launcherIcons[index] = d.launcherIcons[index], d.launcherIcons[index-1]
				d.populateOrderList(list)
			}
		}
		downButton := widget.NewButtonWithIcon("", theme.MoveDownIcon(), nil)
		buttonMap[downButton] = i
		downButton.OnTapped = func() {
			index := buttonMap[downButton]
			if index < len(d.settings.LauncherIcons())-1 {
				d.launcherIcons[index+1], d.launcherIcons[index] = d.launcherIcons[index], d.launcherIcons[index+1]
				d.populateOrderList(list)
			}
		}
		iconRes := appData.Icon(d.settings.IconTheme(), int((float64(d.settings.LauncherIconSize())*d.settings.LauncherZoomScale())*float64(desktop.Instance().Screens().Primary().CanvasScale())))
		icon := widget.NewIcon(iconRes)
		label := widget.NewLabel(appName)
		hbox := widget.NewHBox(upButton, downButton, icon, label)
		list.AddObject(hbox)
	}
	list.Refresh()
}

func (d *settingsUI) loadBarScreen() fyne.CanvasObject {
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

	iconSize := widget.NewEntry()
	iconSize.SetText(fmt.Sprintf("%d", d.settings.LauncherIconSize()))

	zoomScale := widget.NewEntry()
	zoomScale.SetText(fmt.Sprintf("%f", d.settings.LauncherZoomScale()))

	top := widget.NewForm(
		&widget.FormItem{Text: "Launcher Icon Size:", Widget: iconSize},
		&widget.FormItem{Text: "Launcher Zoom Scale:", Widget: zoomScale})

	disableTaskbar := widget.NewCheck("Disable Taskbar", nil)
	disableTaskbar.SetChecked(d.settings.LauncherDisableTaskbar())

	disableZoom := widget.NewCheck("Disable Zoom", nil)
	disableZoom.SetChecked(d.settings.LauncherDisableZoom())

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
			d.settings.setLauncherIconSize(size)

			scale, err := strconv.ParseFloat(zoomScale.Text, 32)
			if err != nil {
				fyne.LogError("Error setting launcher zoom scale", err)
				scale = 2.0
			}
			d.settings.setLauncherZoomScale(scale)
			d.settings.setLauncherDisableTaskbar(disableTaskbar.Checked)
			d.settings.setLauncherDisableZoom(disableZoom.Checked)

			d.settings.setLauncherIcons(d.launcherIcons)
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
	ui := &settingsUI{settings: deskSettings, launcherIcons: deskSettings.LauncherIcons()}

	w := fyne.CurrentApp().NewWindow("FyneDesk Settings")
	fyneSettings := settings.NewSettings()

	tabs := widget.NewTabContainer(
		&widget.TabItem{Text: "Fyne Settings", Icon: theme.FyneLogo(),
			Content: fyneSettings.LoadAppearanceScreen()},
		&widget.TabItem{Text: "Appearance", Icon: fyneSettings.AppearanceIcon(),
			Content: ui.loadAppearanceScreen()},
		&widget.TabItem{Text: "App Bar", Icon: wmtheme.IconifyIcon, Content: ui.loadBarScreen()},
		&widget.TabItem{Text: "Advanced", Icon: theme.SettingsIcon(),
			Content: loadAdvancedScreen()},
	)
	tabs.SetTabLocation(widget.TabLocationLeading)
	w.SetContent(tabs)

	w.Resize(fyne.NewSize(480, 320))
	w.Show()
}
