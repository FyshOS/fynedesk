package icon

import (
	"bytes"
	_ "image/jpeg" // support JPEG images
	"image/png"    // PNG support is required as we use it directly
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jackmordaunt/icns"
	"howett.net/plist"

	"fyne.io/fyne"

	"fyne.io/fynedesk"
	wmtheme "fyne.io/fynedesk/theme"
)

type macOSAppBundle struct {
	DisplayName string `plist:"CFBundleDisplayName"`
	// TODO alternateNames []string
	Executable string `plist:"CFBundleExecutable"`
	runPath    string
	IconFile   string `plist:"CFBundleIconFile"`
	iconPath   string
}

func (m *macOSAppBundle) Name() string {
	return m.DisplayName
}

func (m *macOSAppBundle) Icon(_ string, _ int) fyne.Resource {
	src, err := os.Open(m.iconPath)
	if err != nil {
		fyne.LogError("Failed to read icon data for "+m.iconPath, err)
		return wmtheme.BrokenImageIcon
	}

	icon, err := icns.Decode(src)
	if err != nil {
		fyne.LogError("Failed to parse icon data for "+m.iconPath, err)
		return wmtheme.BrokenImageIcon
	}

	var data bytes.Buffer
	err = png.Encode(&data, icon)
	iconName := filepath.Base(m.iconPath)
	return fyne.NewStaticResource(strings.Replace(iconName, ".icns", ".png", 1), data.Bytes())
}

func (m *macOSAppBundle) Run([]string) error {
	// in macOS test mode we ignore the wm env flags
	return exec.Command("open", "-a", m.runPath).Start()
}

func loadAppBundle(name, path string) fynedesk.AppData {
	buf, err := os.Open(filepath.Join(path, "Contents", "Info.plist"))
	if err != nil {
		fyne.LogError("Unable to read application plist", err)
		return nil
	}

	var data macOSAppBundle
	data.DisplayName = name
	decoder := plist.NewDecoder(buf)
	err = decoder.Decode(&data)
	if err != nil {
		fyne.LogError("Unable to parse application plist", err)
		return nil
	}
	data.runPath = filepath.Join(path, "Contents", "MacOS", data.Executable)

	data.iconPath = filepath.Join(path, "Contents", "Resources", data.IconFile)
	pos := strings.Index(data.iconPath, ".icns")
	if pos == -1 {
		data.iconPath = data.iconPath + ".icns"
	}
	return &data
}

type macOSAppProvider struct {
	rootDirs []string
}

func (m *macOSAppProvider) FindIconsMatchingAppName(theme string, size int, appName string) []fynedesk.AppData {
	panic("implement me")
}

func (m *macOSAppProvider) forEachApplication(f func(string, string) bool) {
	for _, root := range m.rootDirs {
		files, err := ioutil.ReadDir(root)
		if err != nil {
			fyne.LogError("Could not read applications directory "+root, err)
			return
		}
		for _, file := range files {
			if !file.IsDir() || !strings.HasSuffix(file.Name(), ".app") {
				continue // skip non-app bundles
			}
			appDir := filepath.Join(root, file.Name())
			if f(file.Name()[0:len(file.Name())-4], appDir) {
				break
			}
		}
	}
}

func (m *macOSAppProvider) AvailableApps() []fynedesk.AppData {
	var icons []fynedesk.AppData
	m.forEachApplication(func(name, path string) bool {
		app := loadAppBundle(name, path)
		if app != nil {
			icons = append(icons, app)
		}
		return false
	})
	return icons
}

func (m *macOSAppProvider) AvailableThemes() []string {
	//I'm not sure this is relevant on Mac OSX
	return []string{}
}

func (m *macOSAppProvider) FindAppFromName(appName string) fynedesk.AppData {
	var icon fynedesk.AppData
	m.forEachApplication(func(name, path string) bool {
		if name == appName {
			icon = loadAppBundle(name, path)
			if icon != nil {
				return true
			}
		}

		return false
	})

	return icon
}

func (m *macOSAppProvider) FindAppFromWinInfo(win fynedesk.Window) fynedesk.AppData {
	return m.FindAppFromName(win.Title())
}

func (m *macOSAppProvider) DefaultApps() []fynedesk.AppData {
	var apps []fynedesk.AppData

	apps = appendAppIfExists(apps, findOneAppFromNames(m, "Terminal", "iTerm"))
	apps = appendAppIfExists(apps, findOneAppFromNames(m, "Google Chrome", "Firefox", "Safari"))
	apps = appendAppIfExists(apps, findOneAppFromNames(m, "Spark", "AirMail", "Mail"))
	apps = appendAppIfExists(apps, m.FindAppFromName("Photos"))
	apps = appendAppIfExists(apps, m.FindAppFromName("System Preferences"))

	return apps
}

func (m *macOSAppProvider) FindAppsMatching(pattern string) []fynedesk.AppData {
	var icons []fynedesk.AppData
	m.forEachApplication(func(name, path string) bool {
		if !strings.Contains(strings.ToLower(name), strings.ToLower(pattern)) {
			return false
		}

		app := loadAppBundle(name, path)
		if app != nil {
			icons = append(icons, app)
		}
		return false
	})

	return icons
}

// NewMacOSAppProvider creates an instance of an ApplicationProvider that can find and decode macOS apps
func NewMacOSAppProvider() fynedesk.ApplicationProvider {
	return &macOSAppProvider{rootDirs: []string{"/Applications", "/Applications/Utilities",
		"/System/Applications", "/System/Applications/Utilities"}}
}
