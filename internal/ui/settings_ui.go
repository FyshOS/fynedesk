package ui

import (
	"fmt"
	"os/exec"
	"strconv"

	"fyne.io/fyne/canvas"

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
		iconRes := appData.Icon(theme, int((float64(d.settings.LauncherIconSize())*d.settings.LauncherZoomScale())*float64(desktop.Instance().Root().Canvas().Scale())))
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

func (d *settingsUI) populateOrderList(list *widget.Box, add fyne.CanvasObject) {
	var icons []fyne.CanvasObject
	iconSize := desktop.Instance().Settings().LauncherIconSize()
	for i, appName := range d.launcherIcons {
		index := i // capture
		appData := desktop.Instance().IconProvider().FindAppFromName(appName)
		left := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
			d.launcherIcons[index-1], d.launcherIcons[index] = d.launcherIcons[index], d.launcherIcons[index-1]
			d.populateOrderList(list, add)
		})
		if index <= 0 {
			left.Disable()
		}

		remove := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
			if index == 0 {
				d.launcherIcons = d.launcherIcons[1:]
			} else if index == len(d.launcherIcons)-1 {
				d.launcherIcons = d.launcherIcons[:len(d.launcherIcons)-1]
			} else {
				d.launcherIcons = append(d.launcherIcons[:index], d.launcherIcons[index+1])
			}
			d.populateOrderList(list, add)
		})

		right := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
			d.launcherIcons[index+1], d.launcherIcons[index] = d.launcherIcons[index], d.launcherIcons[index+1]
			d.populateOrderList(list, add)
		})
		if index >= len(d.launcherIcons)-1 {
			right.Disable()
		}
		iconRes := appData.Icon(d.settings.IconTheme(), int((float64(d.settings.LauncherIconSize())*d.settings.LauncherZoomScale())*float64(desktop.Instance().Root().Canvas().Scale())))
		icon := canvas.NewImageFromResource(iconRes)
		icon.FillMode = canvas.ImageFillContain
		icon.SetMinSize(fyne.NewSize(iconSize, iconSize))
		label := widget.NewLabelWithStyle(appName, fyne.TextAlignCenter, fyne.TextStyle{})
		hbox := widget.NewVBox(icon, label, widget.NewHBox(left, remove, right))
		icons = append(icons, hbox)
	}

	icons = append(icons, add)
	list.Children = icons
	list.Refresh()
}

func (d *settingsUI) loadBarScreen() fyne.CanvasObject {
	iconWidth := desktop.Instance().Settings().LauncherIconSize()
	addButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {})
	addIcon := canvas.NewImageFromResource(theme.ContentAddIcon())
	addIcon.FillMode = canvas.ImageFillContain
	addIcon.SetMinSize(fyne.NewSize(iconWidth, iconWidth))
	addItem := widget.NewVBox(addIcon, widget.NewLabel("Add Icon"), addButton)
	orderList := widget.NewHBox()
	d.populateOrderList(orderList, addItem)

	addButton.OnTapped = func() {
		newAppPicker("Choose Application", func(data desktop.AppData) {
			d.launcherIcons = append(d.launcherIcons, data.Name())
			d.populateOrderList(orderList, addItem)
		}).Show()
	}

	bar := widget.NewScrollContainer(orderList)

	iconSize := widget.NewEntry()
	iconSize.SetText(fmt.Sprintf("%d", d.settings.LauncherIconSize()))

	zoomScale := widget.NewEntry()
	zoomScale.SetText(fmt.Sprintf("%f", d.settings.LauncherZoomScale()))

	sizeCell := widget.NewHBox(widget.NewLabel("Launcher Icon Size:"), iconSize)
	zoomCell := widget.NewHBox(widget.NewLabel("Launcher Zoom Scale:"), zoomScale)

	disableTaskbar := widget.NewCheck("Disable Taskbar", nil)
	disableTaskbar.SetChecked(d.settings.LauncherDisableTaskbar())

	disableZoom := widget.NewCheck("Disable Zoom", nil)
	disableZoom.SetChecked(d.settings.LauncherDisableZoom())

	details := widget.NewGroup("Configuration",
		fyne.NewContainerWithLayout(layout.NewGridLayout(2),
			sizeCell, zoomCell, disableTaskbar, disableZoom))
	barSettings := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, details, nil, nil),
		bar, details)

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

	header := widget.NewGroup("Icons")
	return fyne.NewContainerWithLayout(layout.NewBorderLayout(header, applyButton, nil, nil),
		header, applyButton, barSettings)
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
