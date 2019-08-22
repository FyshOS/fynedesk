package internal

import (
	"bytes"
	_ "image/jpeg"
	"image/png"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jackmordaunt/icns"
	"howett.net/plist"

	"fyne.io/desktop"
	wmtheme "fyne.io/desktop/theme"
	"fyne.io/fyne"
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

func (m *macOSAppBundle) Run() error {
	return exec.Command("open", "-a", m.runPath).Start()
}

func loadAppBundle(name, path string) desktop.AppData {
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
	rootDir string
}

func (m *macOSAppProvider) FindIconsMatchingAppName(theme string, size int, appName string) []desktop.AppData {
	panic("implement me")
}

func (m *macOSAppProvider) forEachApplication(f func(string, string) bool) {
	files, err := ioutil.ReadDir(m.rootDir)
	if err != nil {
		fyne.LogError("Could not read applications directory "+m.rootDir, err)
		return
	}
	for _, file := range files {
		appDir := filepath.Join(m.rootDir, file.Name())
		if f(file.Name()[0:len(file.Name())-4], appDir) {
			break
		}
	}
}

func (m *macOSAppProvider) FindAppFromName(appName string) desktop.AppData {
	var icon desktop.AppData
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

func (m *macOSAppProvider) FindAppFromWinInfo(win desktop.Window) desktop.AppData {
	return m.FindAppFromName(win.Title())
}

func (m *macOSAppProvider) DefaultApps() []desktop.AppData {
	var apps []desktop.AppData

	apps = appendAppIfExists(apps, findOneAppFromNames(m, "Terminal", "iTerm"))
	apps = appendAppIfExists(apps, findOneAppFromNames(m, "Google Chrome", "Firefox", "Safari"))
	apps = appendAppIfExists(apps, findOneAppFromNames(m, "Spark", "AirMail", "Mail"))
	apps = appendAppIfExists(apps, m.FindAppFromName("Photos"))
	apps = appendAppIfExists(apps, m.FindAppFromName("System Preferences"))

	return apps
}

func (m *macOSAppProvider) FindAppsMatching(pattern string) []desktop.AppData {
	var icons []desktop.AppData
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

func NewMacOSAppProvider() desktop.ApplicationProvider {
	return &macOSAppProvider{rootDir: "/Applications"}
}
