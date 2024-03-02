package ui

import (
	"os"
	"runtime"
	"strings"
	"sync"

	"fyshos.com/fynedesk"

	"fyne.io/fyne/v2"
)

type deskSettings struct {
	background             string
	iconTheme              string
	launcherIcons          []string
	launcherIconSize       float32
	launcherDisableTaskbar bool
	launcherDisableZoom    bool
	launcherZoomScale      float32
	borderButtonPosition   string
	clockFormatting        string

	modifier    fyne.KeyModifier
	moduleNames []string

	narrowPanel, narrowLeftLauncher bool

	listenerLock    sync.Mutex
	changeListeners []chan fynedesk.DeskSettings
}

func (d *deskSettings) Background() string {
	return d.background
}

func (d *deskSettings) IconTheme() string {
	return d.iconTheme
}

func (d *deskSettings) LauncherIcons() []string {
	return d.launcherIcons
}

func (d *deskSettings) LauncherIconSize() float32 {
	return d.launcherIconSize
}

func (d *deskSettings) LauncherDisableTaskbar() bool {
	return d.launcherDisableTaskbar
}

func (d *deskSettings) LauncherDisableZoom() bool {
	return d.launcherDisableZoom
}

func (d *deskSettings) LauncherZoomScale() float32 {
	return d.launcherZoomScale
}

func (d *deskSettings) KeyboardModifier() fyne.KeyModifier {
	return d.modifier
}

func (d *deskSettings) ModuleNames() []string {
	return d.moduleNames
}

func (d *deskSettings) NarrowWidgetPanel() bool {
	return d.narrowPanel
}

func (d *deskSettings) NarrowLeftLauncher() bool {
	return d.narrowLeftLauncher
}

func (d *deskSettings) BorderButtonPosition() string {
	return d.borderButtonPosition
}

func (d *deskSettings) ClockFormatting() string {
	return d.clockFormatting
}

func (d *deskSettings) AddChangeListener(listener chan fynedesk.DeskSettings) {
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

func isModuleEnabled(name string, settings fynedesk.DeskSettings) bool {
	for _, mod := range settings.ModuleNames() {
		if mod == name {
			return true
		}
	}

	return false
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

func (d *deskSettings) setLauncherIconSize(size float32) {
	d.launcherIconSize = size
	fyne.CurrentApp().Preferences().SetInt("launchericonsize", int(d.launcherIconSize))
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

func (d *deskSettings) setLauncherZoomScale(scale float32) {
	d.launcherZoomScale = scale
	fyne.CurrentApp().Preferences().SetFloat("launcherzoomscale", float64(d.launcherZoomScale))
	d.apply()
}

func (d *deskSettings) setKeyboardModifier(mod fyne.KeyModifier) {
	d.modifier = mod
	fyne.CurrentApp().Preferences().SetInt("keyboardmodifier", int(d.modifier))
	d.apply()
}

func (d *deskSettings) setModuleNames(names []string) {
	newModuleNames := strings.Join(names, "|")
	d.moduleNames = names
	fyne.CurrentApp().Preferences().SetString("modulenames", newModuleNames)
	d.apply()
}

func (d *deskSettings) setNarrowLeftLauncher(narrow bool) {
	d.narrowLeftLauncher = narrow
	fyne.CurrentApp().Preferences().SetBool("launchernarrowleft", narrow)
	d.apply()
}

func (d *deskSettings) setNarrowWidgetPanel(narrow bool) {
	d.narrowPanel = narrow
	fyne.CurrentApp().Preferences().SetBool("narrowpanel", narrow)
	d.apply()
}

func (d *deskSettings) setBorderButtonPosition(pos string) {
	d.borderButtonPosition = pos
	fyne.CurrentApp().Preferences().SetString("borderbuttonposition", d.borderButtonPosition)
	d.apply()
}

func (d *deskSettings) setClockFormatting(format string) {
	d.clockFormatting = format
	fyne.CurrentApp().Preferences().SetString("clockformatting", d.clockFormatting)
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
		d.launcherIcons = strings.Split(launcherIcons, "|")
	}
	if len(d.launcherIcons) == 0 {
		defaultApps := fynedesk.Instance().IconProvider().DefaultApps()
		for _, appData := range defaultApps {
			d.launcherIcons = append(d.launcherIcons, appData.Name())
		}
	}

	d.launcherIconSize = float32(fyne.CurrentApp().Preferences().Int("launchericonsize"))
	if d.launcherIconSize == 0 {
		d.launcherIconSize = 48
	}

	d.launcherDisableTaskbar = fyne.CurrentApp().Preferences().Bool("launcherdisabletaskbar")
	d.launcherDisableZoom = fyne.CurrentApp().Preferences().Bool("launcherdisablezoom")

	d.launcherZoomScale = float32(fyne.CurrentApp().Preferences().Float("launcherzoomscale"))
	if d.launcherZoomScale == 0.0 {
		d.launcherZoomScale = 2.0
	}

	defaultModules := "Battery|Brightness|Compositor|Sound|Launcher: Calculate|Launcher: Open URLs|Network|Virtual Desktops|SystemTray"
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" { // testing
		defaultModules = "Battery|Brightness|Sound|Launcher: Calculate|Launcher: Open URLs|Network|Virtual Desktops"
	}
	moduleNames := fyne.CurrentApp().Preferences().StringWithFallback("modulenames", defaultModules)
	if moduleNames != "" {
		d.moduleNames = strings.Split(moduleNames, "|")
	}
	d.modifier = fyne.KeyModifier(fyne.CurrentApp().Preferences().IntWithFallback("keyboardmodifier", int(fyne.KeyModifierSuper)))
	d.narrowLeftLauncher = fyne.CurrentApp().Preferences().BoolWithFallback("launchernarrowleft", true)
	d.narrowPanel = fyne.CurrentApp().Preferences().BoolWithFallback("narrowpanel", true)

	d.borderButtonPosition = fyne.CurrentApp().Preferences().StringWithFallback("borderbuttonposition", "Left")

	d.clockFormatting = fyne.CurrentApp().Preferences().StringWithFallback("clockformatting", "12h")
	d.loadRecents()
}

func (d *deskSettings) loadRecents() {
	str := fyne.CurrentApp().Preferences().String("recentapps")
	desk := fynedesk.Instance().(*desktop)

	var apps []fynedesk.AppData
	list := strings.Split(str, ",")

	for _, s := range list {
		app := desk.icons.FindAppFromName(s)
		if app == nil {
			continue
		}
		apps = append(apps, app)
	}

	desk.recent = apps
}

func (d *deskSettings) saveRecents() {
	var list []string

	for _, a := range fynedesk.Instance().(*desktop).recent {
		list = append(list, a.Name())
	}

	fyne.CurrentApp().Preferences().SetString("recentapps", strings.Join(list, ","))
}

// newDeskSettings loads the user's preferences from environment or config
func newDeskSettings() *deskSettings {
	settings := &deskSettings{}
	settings.load()

	return settings
}
