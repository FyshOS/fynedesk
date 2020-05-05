package ui

import (
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/cmd/fyne_settings/settings"
	"fyne.io/fyne/dialog"
	deskDriver "fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	"fyne.io/fynedesk"
	wmtheme "fyne.io/fynedesk/theme"
	"fyne.io/fynedesk/wm"
)

const randrHelper = "arandr"

type settingsUI struct {
	settings *deskSettings
	win      fyne.Window

	launcherIcons []string
}

func (d *settingsUI) populateThemeIcons(box *fyne.Container, theme string) {
	box.Objects = nil
	for _, appName := range d.launcherIcons {
		appData := fynedesk.Instance().IconProvider().FindAppFromName(appName)
		if appData == nil { // if app was removed!
			continue
		}
		iconRes := appData.Icon(theme, int((float64(d.settings.LauncherIconSize())*d.settings.LauncherZoomScale())*float64(fynedesk.Instance().Screens().Primary().CanvasScale())))
		icon := widget.NewIcon(iconRes)
		box.AddObject(icon)
	}
	box.Refresh()
}

func (d *settingsUI) loadAppearanceScreen() fyne.CanvasObject {
	var bgPathClear *widget.Button
	bgPath := widget.NewEntry()
	bgPath.SetPlaceHolder("Choose an image")
	bgPathClear = widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		bgPath.SetText("")
		bgPathClear.Disable()
	})

	if fyne.CurrentApp().Preferences().String("background") != "" {
		bgPath.SetText(fyne.CurrentApp().Preferences().String("background"))
	} else {
		bgPathClear.Disable()
	}
	bgLabel := widget.NewLabelWithStyle("Background", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	bgButtons := widget.NewHBox(bgPathClear,
		widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
			dialog.ShowFileOpen(func(file fyne.FileReadCloser, err error) {
				if err != nil || file == nil {
					return
				}

				// not advisable for cross-platform but we are desktop only
				path := file.URI()[7:]
				// TODO add a nice preview :)
				_ = file.Close()

				bgPath.SetText(path)
				bgPathClear.Enable()
			}, d.win)
		}))

	themeLabel := widget.NewLabel(d.settings.IconTheme())
	themeIcons := fyne.NewContainerWithLayout(layout.NewHBoxLayout())
	d.populateThemeIcons(themeIcons, d.settings.IconTheme())
	themeList := fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	for _, themeName := range fynedesk.Instance().IconProvider().AvailableThemes() {
		themeButton := widget.NewButton(themeName, nil)
		themeButton.OnTapped = func() {
			themeLabel.SetText(themeButton.Text)
			d.populateThemeIcons(themeIcons, themeButton.Text)
		}
		themeList.AddObject(themeButton)
	}
	top := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, bgLabel, bgButtons),
		bgLabel, bgPath, bgButtons)

	themeFormLabel := widget.NewLabelWithStyle("Icon Theme", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	themeCurrent := widget.NewHBox(layout.NewSpacer(), themeLabel, themeIcons)
	middle := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, themeCurrent, themeFormLabel, nil),
		themeCurrent, themeFormLabel, widget.NewScrollContainer(themeList))

	applyButton := widget.NewHBox(layout.NewSpacer(),
		&widget.Button{Text: "Apply", Style: widget.PrimaryButton, OnTapped: func() {
			d.settings.setBackground(bgPath.Text)
			d.settings.setIconTheme(themeLabel.Text)
		}})

	return fyne.NewContainerWithLayout(layout.NewBorderLayout(top, applyButton, nil, nil), top, applyButton, middle)
}

func (d *settingsUI) populateOrderList(list *widget.Box, add fyne.CanvasObject) {
	var icons []fyne.CanvasObject
	iconSize := fynedesk.Instance().Settings().LauncherIconSize()
	for i, appName := range d.launcherIcons {
		index := i // capture
		appData := fynedesk.Instance().IconProvider().FindAppFromName(appName)
		if appData == nil {
			continue // uninstalled?
		}
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
		iconRes := appData.Icon(d.settings.IconTheme(), int((float64(d.settings.LauncherIconSize())*d.settings.LauncherZoomScale())*float64(fynedesk.Instance().Screens().Primary().CanvasScale())))
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
	iconWidth := fynedesk.Instance().Settings().LauncherIconSize()
	addButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {})
	addIcon := canvas.NewImageFromResource(theme.ContentAddIcon())
	addIcon.FillMode = canvas.ImageFillContain
	addIcon.SetMinSize(fyne.NewSize(iconWidth, iconWidth))
	addItem := widget.NewVBox(addIcon, widget.NewLabel("Add Icon"), addButton)
	orderList := widget.NewHBox()
	d.populateOrderList(orderList, addItem)

	addButton.OnTapped = func() {
		newAppPicker("Choose Application", func(data fynedesk.AppData) {
			d.launcherIcons = append(d.launcherIcons, data.Name())
			d.populateOrderList(orderList, addItem)
		}).Show()
	}

	bar := widget.NewHScrollContainer(orderList)

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
		header, applyButton, widget.NewVBox(bar, details))
}

func (d *settingsUI) loadAdvancedScreen() fyne.CanvasObject {
	var modules []fyne.CanvasObject

	for _, mod := range fynedesk.AvailableModules() {
		name := mod.Name
		enabled := isModuleEnabled(name, d.settings)

		check := widget.NewCheck(name, func(bool) {})
		check.SetChecked(enabled)
		modules = append(modules, check)
	}
	content := fyne.NewContainerWithLayout(layout.NewGridLayout(2), d.loadScreensGroup(),
		widget.NewGroup("Modules", modules...))

	applyButton := widget.NewHBox(layout.NewSpacer(),
		&widget.Button{Text: "Apply", Style: widget.PrimaryButton, OnTapped: func() {
			var names []string
			for _, item := range modules {
				check := item.(*widget.Check)
				if check.Checked {
					names = append(names, check.Text)
				}
			}

			d.settings.setModuleNames(names)
		}})

	return fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, applyButton, nil, nil),
		applyButton, content)
}

func (d *settingsUI) loadKeyboardScreen() fyne.CanvasObject {
	var names, mods, keys []fyne.CanvasObject
	shortcuts := fynedesk.Instance().(wm.ShortcutManager).Shortcuts()
	sort.Slice(shortcuts, func(i, j int) bool {
		return strings.Compare(shortcuts[i].ShortcutName(), shortcuts[j].ShortcutName()) < 0
	})

	for _, shortcut := range shortcuts {
		names = append(names, widget.NewLabel(shortcut.ShortcutName()))
		mods = append(mods, widget.NewLabel(modifierToString(shortcut.Modifier)))
		keys = append(keys, widget.NewLabel(string(shortcut.KeyName)))
	}
	rows := widget.NewHBox(widget.NewGroup("Action", names...),
		widget.NewGroup("Modifier", mods...),
		widget.NewGroup("Key Name", keys...))
	grid := widget.NewScrollContainer(rows)

	applyButton := widget.NewHBox(layout.NewSpacer(),
		&widget.Button{Text: "Apply", Style: widget.PrimaryButton, OnTapped: func() {
		}})

	return fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, applyButton, nil, nil),
		applyButton, grid)
}

func loadScreensTable() fyne.CanvasObject {
	names := widget.NewVBox()
	labels1 := widget.NewVBox()
	values1 := widget.NewVBox()
	labels2 := widget.NewVBox()
	values2 := widget.NewVBox()

	for _, screen := range fynedesk.Instance().Screens().Screens() {
		names.Append(widget.NewLabelWithStyle(screen.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		labels1.Append(widget.NewLabel("Width"))
		values1.Append(widget.NewLabel(fmt.Sprintf("%dpx", screen.Width)))
		labels2.Append(widget.NewLabel("Height"))
		values2.Append(widget.NewLabel(fmt.Sprintf("%dpx", screen.Height)))

		names.Append(widget.NewLabel(""))
		labels1.Append(widget.NewLabel("Scale"))
		values1.Append(widget.NewLabel(fmt.Sprintf("%.1f", screen.Scale)))
		labels2.Append(widget.NewLabel("Applied"))
		values2.Append(widget.NewLabel(fmt.Sprintf("%.1f", screen.CanvasScale())))
	}

	return widget.NewHBox(names, labels1, values1, labels2, values2)
}

func (d *settingsUI) loadScreensGroup() fyne.CanvasObject {
	var displays fyne.CanvasObject
	if _, err := exec.LookPath(randrHelper); err == nil {
		displays = widget.NewButtonWithIcon("Manage Displays", wmtheme.DisplayIcon, func() {
			exec.Command(randrHelper).Start()
		})
	} else {
		displays = widget.NewLabel("This requires " + randrHelper)
	}

	userScale := fyne.CurrentApp().Settings().Scale()
	content := widget.NewVBox(widget.NewLabel(fmt.Sprintf("User scale: %0.2f", userScale)))
	screens := widget.NewGroup("Screens", displays, content)
	screens.Append(loadScreensTable())
	return screens
}

func showSettings(deskSettings *deskSettings) {
	ui := &settingsUI{settings: deskSettings, launcherIcons: deskSettings.LauncherIcons()}

	w := fyne.CurrentApp().NewWindow("FyneDesk Settings")
	ui.win = w
	fyneSettings := settings.NewSettings()

	tabs := widget.NewTabContainer(
		&widget.TabItem{Text: "Fyne Settings", Icon: theme.FyneLogo(),
			Content: fyneSettings.LoadAppearanceScreen(w)},
		&widget.TabItem{Text: "Appearance", Icon: fyneSettings.AppearanceIcon(),
			Content: ui.loadAppearanceScreen()},
		&widget.TabItem{Text: "App Bar", Icon: wmtheme.IconifyIcon, Content: ui.loadBarScreen()},
		&widget.TabItem{Text: "Keyboard", Icon: wmtheme.KeyboardIcon, Content: ui.loadKeyboardScreen()},
		&widget.TabItem{Text: "Advanced", Icon: theme.SettingsIcon(),
			Content: ui.loadAdvancedScreen()},
	)
	tabs.SetTabLocation(widget.TabLocationLeading)
	w.SetContent(tabs)

	w.Resize(fyne.NewSize(480, 320))
	w.Show()
}

func modifierToString(mods deskDriver.Modifier) string {
	var s []string
	if (mods & deskDriver.ShiftModifier) != 0 {
		s = append(s, "Shift")
	}
	if (mods & deskDriver.ControlModifier) != 0 {
		s = append(s, "Control")
	}
	if (mods & deskDriver.AltModifier) != 0 {
		s = append(s, "Alt")
	}
	if (mods & deskDriver.SuperModifier) != 0 {
		if runtime.GOOS == "darwin" {
			s = append(s, "Command")
		} else {
			s = append(s, "Super")
		}
	}
	return strings.Join(s, "+")
}
