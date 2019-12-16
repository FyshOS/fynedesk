package desktop

import (
	"fmt"
	wmtheme "fyne.io/desktop/theme"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/cmd/fyne_settings/settings"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"os"
	"os/exec"
	"strings"

	"fyne.io/fyne"
)

// DeskSettings describes the configuration options available for Fyne desktop
type DeskSettings interface {
	Background() string
	IconTheme() string
	DefaultApps() string
}

type deskSettings struct {
	background  string
	iconTheme   string
	defaultApps string
}

const randrHelper = "arandr"

func (d *deskSettings) Background() string {
	return d.background
}

func (d *deskSettings) IconTheme() string {
	return d.iconTheme
}

func (d *deskSettings) DefaultApps() string {
	return d.defaultApps
}

func (d *deskSettings) setBackground(name string) {
	d.background = name
	fyne.CurrentApp().Preferences().SetString("background", d.background)
	canvas.Refresh(Instance().(*deskLayout).container)
}

func (d *deskSettings) setIconTheme(name string) {
	d.iconTheme = name
	fyne.CurrentApp().Preferences().SetString("icontheme", d.iconTheme)
	canvas.Refresh(Instance().(*deskLayout).container)
}

func (d *deskSettings) setDefaultApps(defaultApps []string) {
	var newDefaultApps string
	for i, app := range defaultApps {
		fmt.Println(app)
		if i == 0 {
			newDefaultApps = app
		} else {
			newDefaultApps += "|" + app
		}
	}
	fyne.CurrentApp().Preferences().SetString("defaultapps", newDefaultApps)
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

	d.defaultApps = fyne.CurrentApp().Preferences().String("defaultapps")

	if d.defaultApps == "" {
		for i, app := range Instance().IconProvider().DefaultApps() {
			if i == 0 {
				d.defaultApps = app.Name()
			} else {
				d.defaultApps += "|" + app.Name()
			}
		}
	}
}

func (d *deskSettings) loadAppearanceScreen() fyne.CanvasObject {
	bgEntry := widget.NewEntry()
	if fyne.CurrentApp().Preferences().String("background") == "" {
		bgEntry.SetText("Default")
	} else {
		bgEntry.SetText(fyne.CurrentApp().Preferences().String("background"))
	}

	iconThemes := widget.NewSelect(Instance().IconProvider().AvailableThemes(), d.setIconTheme)
	iconThemes.SetSelected(d.iconTheme)

	button := &widget.Button{Text: "Apply", Style: widget.PrimaryButton, OnTapped: func() { d.setBackground(bgEntry.Text) }}

	top := widget.NewForm(
		&widget.FormItem{Text: "Background", Widget: bgEntry},
		&widget.FormItem{Text: "Icon Theme", Widget: iconThemes})
	bottom := widget.NewHBox(layout.NewSpacer(), button)

	return widget.NewVBox(top, bottom)
}

func (d *deskSettings) listAppMatches(list *fyne.Container, input string) {
	dataRange := Instance().IconProvider().FindAppsMatching(input)
	defaultApps := strings.SplitN(Instance().Settings().DefaultApps(), "|", -1)
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
		}
		list.AddObject(hbox)
	}
}

func (d *deskSettings) loadBarScreen() fyne.CanvasObject {
	list := fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	entry := widget.NewEntry()
	entry.SetPlaceHolder("Start Typing An Application Name")
	entry.OnChanged = func(input string) {
		list.Objects = nil
		if input == "" {
			return
		}

		d.listAppMatches(list, input)
	}
	barSettings := fyne.NewContainerWithLayout(layout.NewBorderLayout(entry, nil, nil, nil), entry, widget.NewScrollContainer(list))

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
