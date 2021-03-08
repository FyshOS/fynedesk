package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/cmd/fyne_settings/settings"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	deskDriver "fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

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
		iconRes := appData.Icon(theme, int((d.settings.LauncherIconSize()*d.settings.LauncherZoomScale())*fynedesk.Instance().Screens().Primary().CanvasScale()))
		icon := widget.NewIcon(iconRes)
		box.Add(icon)
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
	bgDialog := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
		if err != nil || file == nil {
			return
		}

		// not advisable for cross-platform but we are desktop only
		path := file.URI().String()[7:]
		// TODO add a nice preview :)
		_ = file.Close()

		bgPath.SetText(path)
		bgPathClear.Enable()
	}, d.win)
	bgDialog.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".jpeg", ".png", ".svg"}))
	if dir, err := getPicturesDir(); err == nil {
		bgDialog.SetLocation(dir)
	} else {
		fyne.LogError("error finding pictures dir, falling back to home directory", err)
	}

	bgButtons := container.NewHBox(bgPathClear,
		widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
			bgDialog.Show()
		}))

	clockLabel := widget.NewLabelWithStyle("Clock Format", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	clockFormat := &widget.RadioGroup{Options: []string{"12h", "24h"}, Required: true, Horizontal: true}
	clockFormat.SetSelected(d.settings.ClockFormatting())

	themeLabel := widget.NewLabel(d.settings.IconTheme())
	themeIcons := container.NewHBox()
	d.populateThemeIcons(themeIcons, d.settings.IconTheme())
	themeList := container.NewVBox()
	for _, themeName := range fynedesk.Instance().IconProvider().AvailableThemes() {
		themeButton := widget.NewButton(themeName, nil)
		themeButton.OnTapped = func() {
			themeLabel.SetText(themeButton.Text)
			d.populateThemeIcons(themeIcons, themeButton.Text)
		}
		themeList.Add(themeButton)
	}

	bg := container.NewBorder(nil, nil, bgLabel, bgButtons, bgPath)
	time := container.NewBorder(nil, nil, clockLabel, clockFormat)
	top := container.NewVBox(bg, time)

	themeFormLabel := widget.NewLabelWithStyle("Icon Theme", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	themeCurrent := container.NewHBox(layout.NewSpacer(), themeLabel, themeIcons)
	bottom := container.NewBorder(nil, themeCurrent, themeFormLabel, nil, container.NewScroll(themeList))

	applyButton := container.NewHBox(layout.NewSpacer(),
		&widget.Button{Text: "Apply", Importance: widget.HighImportance, OnTapped: func() {
			d.settings.setBackground(bgPath.Text)
			d.settings.setIconTheme(themeLabel.Text)
			d.settings.setClockFormatting(clockFormat.Selected)
		}})

	return container.NewBorder(top, applyButton, nil, nil, bottom)
}

func (d *settingsUI) populateOrderList(list *fyne.Container, add fyne.CanvasObject) {
	var icons []fyne.CanvasObject
	iconSize := float32(fynedesk.Instance().Settings().LauncherIconSize())
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
		iconRes := appData.Icon(d.settings.IconTheme(), int((d.settings.LauncherIconSize()*d.settings.LauncherZoomScale())*fynedesk.Instance().Screens().Primary().CanvasScale()))
		icon := canvas.NewImageFromResource(iconRes)
		icon.FillMode = canvas.ImageFillContain
		icon.SetMinSize(fyne.NewSize(iconSize, iconSize))
		label := widget.NewLabelWithStyle(appName, fyne.TextAlignCenter, fyne.TextStyle{})
		hbox := container.NewVBox(icon, label, container.NewHBox(left, remove, right))
		icons = append(icons, hbox)
	}

	icons = append(icons, add)
	list.Objects = icons
	list.Refresh()
}

func (d *settingsUI) loadBarScreen() fyne.CanvasObject {
	iconWidth := float32(fynedesk.Instance().Settings().LauncherIconSize())
	addButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {})
	addIcon := canvas.NewImageFromResource(theme.ContentAddIcon())
	addIcon.FillMode = canvas.ImageFillContain
	addIcon.SetMinSize(fyne.NewSize(iconWidth, iconWidth))
	addItem := container.NewVBox(addIcon, widget.NewLabel("Add Icon"), addButton)
	orderList := container.NewHBox()
	d.populateOrderList(orderList, addItem)

	addButton.OnTapped = func() {
		newAppPicker("Choose Application", func(data fynedesk.AppData) {
			d.launcherIcons = append(d.launcherIcons, data.Name())
			d.populateOrderList(orderList, addItem)
		}).Show()
	}

	bar := container.NewHScroll(orderList)

	iconSize := widget.NewEntry()
	iconSize.Wrapping = fyne.TextWrapOff
	iconSize.SetText(fmt.Sprintf("%0.0f", d.settings.LauncherIconSize()))

	zoomScale := widget.NewEntry()
	zoomScale.Wrapping = fyne.TextWrapOff
	zoomScale.SetText(fmt.Sprintf("%0.2f", d.settings.LauncherZoomScale()))

	sizeCell := container.NewHBox(widget.NewLabel("Launcher Icon Size:"), iconSize)
	zoomCell := container.NewHBox(widget.NewLabel("Launcher Zoom Scale:"), zoomScale)

	disableTaskbar := widget.NewCheck("Disable Taskbar", nil)
	disableTaskbar.SetChecked(d.settings.LauncherDisableTaskbar())

	disableZoom := widget.NewCheck("Disable Zoom", nil)
	disableZoom.SetChecked(d.settings.LauncherDisableZoom())

	details := widget.NewCard("Configuration", "",
		container.NewGridWithColumns(2, sizeCell, zoomCell, disableTaskbar, disableZoom))

	applyButton := container.NewHBox(layout.NewSpacer(),
		&widget.Button{Text: "Apply", Importance: widget.HighImportance, OnTapped: func() {
			size, err := strconv.Atoi(iconSize.Text)
			if err != nil {
				fyne.LogError("error setting launcher icon size", err)
				size = 32
			}
			d.settings.setLauncherIconSize(float32(size))

			scale, err := strconv.ParseFloat(zoomScale.Text, 32)
			if err != nil {
				fyne.LogError("Error setting launcher zoom scale", err)
				scale = 2.0
			}
			d.settings.setLauncherZoomScale(float32(scale))
			d.settings.setLauncherDisableTaskbar(disableTaskbar.Checked)
			d.settings.setLauncherDisableZoom(disableZoom.Checked)

			d.settings.setLauncherIcons(d.launcherIcons)
		}})

	return container.NewBorder(nil, applyButton, nil, nil,
		widget.NewCard("Icons", "", container.NewVBox(bar, details)))
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
	content := container.NewGridWithColumns(2, d.loadScreensGroup(),
		widget.NewCard("Modules", "", container.NewVBox(modules...)))

	applyButton := container.NewHBox(layout.NewSpacer(),
		&widget.Button{Text: "Apply", Importance: widget.HighImportance, OnTapped: func() {
			var names []string
			for _, item := range modules {
				check := item.(*widget.Check)
				if check.Checked {
					names = append(names, check.Text)
				}
			}

			d.settings.setModuleNames(names)
		}})

	return container.NewBorder(nil, applyButton, nil, nil, content)
}

func (d *settingsUI) loadKeyboardScreen() fyne.CanvasObject {
	var names, mods, keys []fyne.CanvasObject
	shortcuts := fynedesk.Instance().(wm.ShortcutManager).Shortcuts()
	sort.Slice(shortcuts, func(i, j int) bool {
		return strings.Compare(shortcuts[i].ShortcutName(), shortcuts[j].ShortcutName()) < 0
	})

	for _, shortcut := range shortcuts {
		names = append(names, widget.NewLabel(shortcut.ShortcutName()))
		mods = append(mods, widget.NewLabel(modifierToString(shortcut.Modifier, d.settings.modifier)))
		keys = append(keys, widget.NewLabel(string(shortcut.KeyName)))
	}
	modVBox := container.NewVBox(mods...)
	rows := container.NewHBox(widget.NewCard("Action", "", container.NewVBox(names...)),
		widget.NewCard("Modifier", "", modVBox),
		widget.NewCard("Key Name", "", container.NewVBox(keys...)))
	grid := container.NewScroll(rows)

	userMod := d.settings.modifier
	modType := widget.NewRadioGroup([]string{"Super", "Alt"}, func(mod string) {
		if mod == "Alt" {
			userMod = deskDriver.AltModifier
		} else {
			userMod = deskDriver.SuperModifier
		}

		var mods []fyne.CanvasObject
		for _, shortcut := range shortcuts {
			mods = append(mods, widget.NewLabel(modifierToString(shortcut.Modifier, userMod)))
		}
		modVBox.Objects = mods
		modVBox.Refresh()
	})
	modType.Horizontal = true
	if d.settings.modifier == deskDriver.AltModifier {
		modType.Selected = "Alt"
	} else {
		modType.Selected = "Super"
	}

	applyButton := container.NewHBox(layout.NewSpacer(),
		&widget.Button{Text: "Apply", Importance: widget.HighImportance, OnTapped: func() {
			d.settings.setKeyboardModifier(userMod)
		}})

	return container.NewBorder(container.NewHBox(widget.NewLabel("Preferred modifier key: "), modType),
		applyButton, nil, nil, grid)
}

func loadScreensTable() fyne.CanvasObject {
	names := container.NewVBox()
	labels1 := container.NewVBox()
	values1 := container.NewVBox()
	labels2 := container.NewVBox()
	values2 := container.NewVBox()

	for _, screen := range fynedesk.Instance().Screens().Screens() {
		names.Add(widget.NewLabelWithStyle(screen.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		labels1.Add(widget.NewLabel("Width"))
		values1.Add(widget.NewLabel(fmt.Sprintf("%dpx", screen.Width)))
		labels2.Add(widget.NewLabel("Height"))
		values2.Add(widget.NewLabel(fmt.Sprintf("%dpx", screen.Height)))

		names.Add(widget.NewLabel(""))
		labels1.Add(widget.NewLabel("Scale"))
		values1.Add(widget.NewLabel(fmt.Sprintf("%.1f", screen.Scale)))
		labels2.Add(widget.NewLabel("Applied"))
		values2.Add(widget.NewLabel(fmt.Sprintf("%.1f", screen.CanvasScale())))
	}

	return container.NewHBox(names, labels1, values1, labels2, values2)
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
	content := container.NewVBox(widget.NewLabel(fmt.Sprintf("User scale: %0.2f", userScale)))
	screens := widget.NewCard("Screens", "", container.NewVBox(displays, content, loadScreensTable()))
	return screens
}

func showSettings(deskSettings *deskSettings) {
	ui := &settingsUI{settings: deskSettings, launcherIcons: deskSettings.LauncherIcons()}

	w := fyne.CurrentApp().NewWindow("FyneDesk Settings")
	ui.win = w
	fyneSettings := settings.NewSettings()

	tabs := container.NewAppTabs(
		&container.TabItem{Text: "Fyne Settings", Icon: theme.FyneLogo(),
			Content: fyneSettings.LoadAppearanceScreen(w)},
		&container.TabItem{Text: "Appearance", Icon: fyneSettings.AppearanceIcon(),
			Content: ui.loadAppearanceScreen()},
		&container.TabItem{Text: "App Bar", Icon: wmtheme.IconifyIcon, Content: ui.loadBarScreen()},
		&container.TabItem{Text: "Keyboard", Icon: wmtheme.KeyboardIcon, Content: ui.loadKeyboardScreen()},
		&container.TabItem{Text: "Advanced", Icon: theme.SettingsIcon(),
			Content: ui.loadAdvancedScreen()},
	)
	tabs.SetTabLocation(container.TabLocationLeading)
	w.SetContent(tabs)

	w.Resize(fyne.NewSize(480, 320))
	w.Show()
}

func modifierToString(mods deskDriver.Modifier, userMod deskDriver.Modifier) string {
	var s []string
	if (mods & fynedesk.UserModifier) != 0 {
		mods |= userMod
	}

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

func getPicturesDir() (fyne.ListableURI, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	const xdg = "xdg-user-dir"
	if _, err := exec.LookPath(xdg); err == nil {
		cmd := exec.Command(xdg, "PICTURES")

		out, err := cmd.Output()
		location := string(out[:len(out)-1]) // Remove \n at the end
		if err == nil && location != home {
			uri := storage.NewFileURI(location)
			return storage.ListerForURI(uri)
		}
	}

	uri, err := storage.Child(storage.NewFileURI(home), "Pictures")
	if err != nil {
		return nil, err
	}

	return storage.ListerForURI(uri)
}
