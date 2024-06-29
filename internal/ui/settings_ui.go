package ui

import (
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/cmd/fyne_settings/settings"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fyshos.com/fynedesk"
	wmtheme "fyshos.com/fynedesk/theme"
	"fyshos.com/fynedesk/wm"
)

const randrHelper = "arandr"

//go:embed "themes/*"
var bundledThemes embed.FS

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

	layoutLabel := widget.NewLabelWithStyle("Layout", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	narrowBar := widget.NewCheck("Narrow Side App Bar", nil)
	narrowBar.Checked = d.settings.NarrowLeftLauncher()
	narrowWidget := widget.NewCheck("Narrow Widget Bar", nil)
	narrowWidget.Checked = d.settings.NarrowWidgetPanel()

	borderButtonLabel := widget.NewLabelWithStyle("Border Button Position", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	borderButton := &widget.Select{Options: []string{"Left", "Right"}}
	borderButton.SetSelected(d.settings.BorderButtonPosition())

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
	lay := container.NewBorder(nil, nil, layoutLabel,
		container.NewGridWithColumns(2, narrowBar, narrowWidget))
	border := container.NewBorder(nil, nil, borderButtonLabel, borderButton)
	top := container.NewVBox(bg, time, lay, border)

	themeFormLabel := widget.NewLabelWithStyle("Icon Theme", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	themeCurrent := container.NewHBox(layout.NewSpacer(), themeLabel, themeIcons)
	bottom := container.NewBorder(nil, themeCurrent, themeFormLabel, nil, container.NewScroll(themeList))

	applyButton := container.NewHBox(layout.NewSpacer(),
		&widget.Button{Text: "Apply", Importance: widget.HighImportance, OnTapped: func() {
			d.settings.setBackground(bgPath.Text)
			d.settings.setIconTheme(themeLabel.Text)
			d.settings.setClockFormatting(clockFormat.Selected)
			d.settings.setBorderButtonPosition(borderButton.Selected)
			d.settings.setNarrowLeftLauncher(narrowBar.Checked)
			d.settings.setNarrowWidgetPanel(narrowWidget.Checked)
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
	iconSize.SetText(strconv.FormatFloat(float64(d.settings.LauncherIconSize()), 'f', 0, 32))

	zoomScale := widget.NewEntry()
	zoomScale.Wrapping = fyne.TextWrapOff
	zoomScale.SetText(strconv.FormatFloat(float64(d.settings.LauncherZoomScale()), 'f', 2, 64))

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
	content := container.NewHBox(d.loadScreensGroup(),
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
			userMod = fyne.KeyModifierAlt
		} else {
			userMod = fyne.KeyModifierSuper
		}

		var mods []fyne.CanvasObject
		for _, shortcut := range shortcuts {
			mods = append(mods, widget.NewLabel(modifierToString(shortcut.Modifier, userMod)))
		}
		modVBox.Objects = mods
		modVBox.Refresh()
	})
	modType.Horizontal = true
	if d.settings.modifier == fyne.KeyModifierAlt {
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
	labels1 := container.NewVBox()
	values1 := container.NewVBox()
	labels2 := container.NewVBox()
	values2 := container.NewVBox()

	all := container.NewVBox()
	for _, screen := range fynedesk.Instance().Screens().Screens() {
		all.Add(widget.NewLabelWithStyle(screen.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		labels1.Add(widget.NewLabel("Width"))
		values1.Add(widget.NewLabel(strconv.Itoa(screen.Width) + "px"))
		labels2.Add(widget.NewLabel("Height"))
		values2.Add(widget.NewLabel(strconv.Itoa(screen.Height) + "px"))

		labels1.Add(widget.NewLabel("Scale"))
		values1.Add(widget.NewLabel(strconv.FormatFloat(float64(screen.Scale), 'f', 1, 32)))
		labels2.Add(widget.NewLabel("Applied"))
		values2.Add(widget.NewLabel(strconv.FormatFloat(float64(screen.CanvasScale()), 'f', 1, 32)))
		all.Add(container.NewHBox(labels1, values1, labels2, values2))
	}

	return all
}

func (d *settingsUI) loadScreensGroup() fyne.CanvasObject {
	var displays fyne.CanvasObject
	if _, err := exec.LookPath(randrHelper); err == nil {
		displays = widget.NewButtonWithIcon("Manage Displays", wmtheme.DisplayIcon, func() {
			e := exec.Command(randrHelper).Start()
			if e != nil {
				fyne.LogError("", e)
			}
		})
	} else {
		displays = widget.NewLabel("This requires " + randrHelper)
	}

	userScale := fyne.CurrentApp().Settings().Scale()
	if userScale == 0.0 {
		userScale = 1.0
	}
	content := container.NewVBox(widget.NewLabel("User scale: " + strconv.FormatFloat(float64(userScale), 'f', 2, 32)))
	screens := widget.NewCard("Screens", "", container.NewVBox(displays, content, loadScreensTable()))
	return screens
}

func (d *settingsUI) loadThemeScreen() fyne.CanvasObject {
	var themeList []string

	embedList, _ := bundledThemes.ReadDir("themes")
	for _, dir := range embedList {
		themeList = append(themeList, dir.Name())
	}

	storageRoot := fyne.CurrentApp().Storage().RootURI()
	themes, _ := storage.Child(storageRoot, "themes")
	list, err := storage.List(themes)
	if err != nil {
		fyne.LogError("Unable to list themes - missing?", err)
		themeList = make([]string, 1)
	} else {
		for _, l := range list {
			if false {
				themeList = append(themeList, l.Name())
			}
		}
	}

	useTheme := func(name string) {
		dest := filepath.Join(filepath.Dir(storageRoot.Path()), "theme.json")
		out, _ := os.Create(dest)
		defer out.Close()
		if name == "default" {
			_, _ = io.WriteString(out, "{}")
			return
		}

		var in io.ReadCloser
		if builtin, err := bundledThemes.Open(filepath.Join("themes/", name, "theme.json")); err == nil {
			in = builtin
		} else {
			source := filepath.Join(themes.Path(), name, "theme.json")
			in, _ = os.Open(source)
		}
		defer in.Close()

		_, err = io.Copy(out, in)
	}
	return widget.NewList(
		func() int {
			return len(themeList)
		},
		func() fyne.CanvasObject {
			install := widget.NewButtonWithIcon("Install", theme.ComputerIcon(), nil)
			preview := &canvas.Image{FillMode: canvas.ImageFillContain}
			preview.SetMinSize(fyne.NewSize(160, 90))
			return container.NewBorder(nil, nil, nil, preview,
				container.NewBorder(nil, install, nil, nil,
					widget.NewRichTextFromMarkdown("## Theme Name\n\nDescription...")))
		},
		func(id widget.ListItemID, o fyne.CanvasObject) {
			outer := o.(*fyne.Container)
			inner := outer.Objects[0].(*fyne.Container)
			b := inner.Objects[1].(*widget.Button)
			b.OnTapped = func() {
				useTheme(themeList[id])
			}
			p := outer.Objects[1].(*canvas.Image)
			if builtin, err := bundledThemes.Open(filepath.Join("themes/", themeList[id], "preview.png")); err == nil {
				data, _ := io.ReadAll(builtin)
				p.Resource = fyne.NewStaticResource(themeList[id]+"/preview.json", data)
				p.File = ""
				_ = builtin.Close()
			} else {
				source := filepath.Join(themes.Path(), themeList[id], "preview.png")
				p.File = source
				p.Resource = nil
			}
			p.Refresh()

			l := inner.Objects[0].(*widget.RichText)
			title := cases.Title(language.Make("en")).String(themeList[id])
			l.ParseMarkdown(fmt.Sprintf("## %s\n\nDescription...", title))
		})
}

func (w *widgetPanel) showSettings() {
	if w.settings != nil {
		w.settings.CenterOnScreen()
		w.settings.Show()

		for _, win := range w.desk.WindowManager().Windows() {
			if win.Properties().Title() == w.settings.Title() {
				w.desk.WindowManager().RaiseToTop(win)
				break
			}
		}
		return
	}

	deskSettings := w.desk.Settings().(*deskSettings)
	ui := &settingsUI{
		settings:      deskSettings,
		launcherIcons: deskSettings.LauncherIcons(),
	}

	win := fyne.CurrentApp().NewWindow("FyneDesk Settings")
	ui.win = win
	fyneSettings := settings.NewSettings()

	tabs := container.NewAppTabs(
		&container.TabItem{Text: "Fyne Settings", Icon: wmtheme.FyneLogo,
			Content: fyneSettings.LoadAppearanceScreen(win)},
		&container.TabItem{Text: "Appearance", Icon: fyneSettings.AppearanceIcon(),
			Content: ui.loadAppearanceScreen()},
		&container.TabItem{Text: "Theme", Icon: theme.ColorPaletteIcon(), Content: ui.loadThemeScreen()},
		&container.TabItem{Text: "App Bar", Icon: wmtheme.IconifyIcon, Content: ui.loadBarScreen()},
		&container.TabItem{Text: "Keyboard", Icon: wmtheme.KeyboardIcon, Content: ui.loadKeyboardScreen()},
		&container.TabItem{Text: "Advanced", Icon: theme.SettingsIcon(),
			Content: ui.loadAdvancedScreen()},
	)
	tabs.SetTabLocation(container.TabLocationLeading)
	win.SetContent(tabs)
	win.Resize(fyne.NewSize(480, 320))

	win.SetCloseIntercept(func() {
		win.Hide()
	})
	w.settings = win
	win.Show()
}

func modifierToString(mods fyne.KeyModifier, userMod fyne.KeyModifier) string {
	var s []string
	if (mods & fynedesk.UserModifier) != 0 {
		mods |= userMod
	}

	if (mods & fyne.KeyModifierShift) != 0 {
		s = append(s, "Shift")
	}
	if (mods & fyne.KeyModifierControl) != 0 {
		s = append(s, "Control")
	}
	if (mods & fyne.KeyModifierAlt) != 0 {
		s = append(s, "Alt")
	}
	if (mods & fyne.KeyModifierSuper) != 0 {
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
