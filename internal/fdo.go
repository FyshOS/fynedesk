package internal // import "fyne.io/desktop/internal"

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/desktop"
	wmTheme "fyne.io/desktop/theme"
	"fyne.io/fyne"
)

var iconExtensions = []string{".png", ".svg"}

//fdoApplicationData is a structure that contains information about .desktop files
type fdoApplicationData struct {
	name     string // Application name
	iconName string // Icon name
	iconPath string // Icon path
	exec     string // Command to execute application
}

//Name returns the name associated with an fdo app
func (data *fdoApplicationData) Name() string {
	return data.name
}

//IconName returns the name of the icon that an fdo app wishes to use
func (data *fdoApplicationData) IconName() string {
	return data.iconName
}

//IconPath returns the path of the icon that an fdo app wishes to use
func (data *fdoApplicationData) Icon(theme string, size int) fyne.Resource {
	path := data.iconPath
	if path == "" {
		path = fdoLookupIconPath(theme, size, data.iconName)
		if path == "" {
			return wmTheme.BrokenImageIcon
		}
	}
	return loadIcon(path)
}

//extractArgs sanitises argument parameters from an Exec configuration
func extractArgs(args []string) []string {
	var ret []string
	for _, arg := range args {
		if len(arg) >= 2 && arg[0] == '%' {
			continue
		}
		ret = append(ret, arg)
	}

	return ret
}

//Run executes the command for this fdo app
func (data *fdoApplicationData) Run(env []string) error {
	vars := os.Environ()
	for _, e := range env {
		vars = append(vars, e)
	}
	commands := strings.Split(data.exec, " ")
	command := commands[0]
	if command[0] == '"' {
		command = command[1 : len(command)-1]
	}

	cmd := exec.Command(command)
	if len(commands) > 1 {
		cmd.Args = extractArgs(commands[1:])
	}

	cmd.Env = vars
	log.Println("env", vars)
	return cmd.Start()
}

func loadIcon(path string) fyne.Resource {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fyne.LogError("Failed to load image", err)
		return nil
	}

	return fyne.NewStaticResource(filepath.Base(path), data)
}

//fdoLookupXdgDataDirs returns a string slice of all XDG_DATA_DIRS
func fdoLookupXdgDataDirs() []string {
	dataLocation := os.Getenv("XDG_DATA_DIRS")
	locationLookup := strings.Split(dataLocation, ":")
	if len(locationLookup) == 0 || (len(locationLookup) == 1 && locationLookup[0] == "") {
		var fallbackLocations []string
		fallbackLocations = append(fallbackLocations, "/usr/local/share")
		fallbackLocations = append(fallbackLocations, "/usr/share")
		return fallbackLocations
	}
	return locationLookup
}

func fdoForEachApplicationFile(f func(data desktop.AppData) bool) {
	locationLookup := fdoLookupXdgDataDirs()
	for _, dataDir := range locationLookup {
		testLocation := filepath.Join(dataDir, "applications")
		files, err := ioutil.ReadDir(testLocation)
		if err != nil {
			continue
		}
		for _, file := range files {
			if strings.HasPrefix(file.Name(), ".") || file.IsDir() {
				continue
			}

			icon := newFdoIconData(filepath.Join(testLocation, file.Name()))
			if icon == nil {
				continue
			}

			if f(icon) {
				return
			}
		}
	}
}

//fdoLookupApplicationByMetadata looks up an application by comparing the requested name to the contents of .desktop files
func fdoLookupApplicationByMetadata(appName string) desktop.AppData {
	var returnIcon desktop.AppData
	fdoForEachApplicationFile(func(icon desktop.AppData) bool {
		if icon.(*fdoApplicationData).name == appName || icon.(*fdoApplicationData).exec == appName {
			returnIcon = icon
			return true
		}
		return false
	})
	return returnIcon
}

//fdoLookupApplication looks up an application by name and returns an fdoApplicationData struct
func fdoLookupApplication(appName string) desktop.AppData {
	locationLookup := fdoLookupXdgDataDirs()
	for _, dataDir := range locationLookup {
		testLocation := filepath.Join(dataDir, "applications", appName+".desktop")
		if _, err := os.Stat(testLocation); err == nil {
			return newFdoIconData(testLocation)
		}
	}
	//If no match was found checking by filenames, check by file contents
	return fdoLookupApplicationByMetadata(appName)
}

//fdoLookupApplicationPartial looks up an application by a partial name and returns all matches in an fdoApplicationData struct slice
func fdoLookupApplicationsMatching(appName string) []desktop.AppData {
	var icons []desktop.AppData
	fdoForEachApplicationFile(func(icon desktop.AppData) bool {
		if icon == nil {
			return false
		}
		if strings.Contains(strings.ToLower(icon.(*fdoApplicationData).name), strings.ToLower(appName)) ||
			strings.Contains(strings.ToLower(icon.(*fdoApplicationData).exec), strings.ToLower(appName)) {
			icons = append(icons, icon)
		}

		return false
	})

	return icons
}

func fdoLookupApplications() []desktop.AppData {
	var icons []desktop.AppData
	fdoForEachApplicationFile(func(icon desktop.AppData) bool {
		if icon == nil {
			return false
		}
		icons = append(icons, icon)
		return false
	})
	return icons
}

//fdoLookupApplicationWinInfo looks up an application based on window info and returns an fdoApplicationData struct
func fdoLookupApplicationWinInfo(win desktop.Window) desktop.AppData {
	icon := fdoLookupApplication(win.Title())
	if icon != nil {
		return icon
	}
	for _, class := range win.Class() {
		icon := fdoLookupApplication(class)
		if icon != nil {
			return icon
		}
	}
	icon = fdoLookupApplication(win.Command())
	if icon != nil {
		return icon
	}
	icon = fdoLookupApplication(win.IconName())
	if icon != nil {
		return icon
	}
	return &fdoApplicationData{name: win.Title()}
}

// lookupAnyIconSizeInThemeDir walks file inside of a directory in reverse order to make sure larger sized icons are found first
func lookupAnyIconSizeInThemeDir(directory string, joiner string, iconName string) string {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return ""
	}
	// Lets walk the files in reverse so bigger icons are selected first (Unless it is a 3 digit icon size like 128)
	// Example is /usr/share/icons/icon_theme/<size>/<joiner>/xterm.png
	for i := len(files) - 1; i >= 0; i-- {
		f := files[i]
		if strings.HasPrefix(f.Name(), ".") || !f.IsDir() {
			continue
		}
		matchDir := filepath.Join(directory, f.Name())
		for _, extension := range iconExtensions {
			testIcon := filepath.Join(matchDir, joiner, iconName+extension)
			if _, err := os.Stat(testIcon); err == nil {
				return testIcon
			}
		}
	}

	directory = filepath.Join(directory, joiner)
	files, err = ioutil.ReadDir(directory)
	if err != nil {
		return ""
	}
	// And then walk the directories in <joiner>/<size> order
	// Example is /usr/share/icons/icon_theme/<joiner>/<size>/xterm.png
	for i := len(files) - 1; i >= 0; i-- {
		f := files[i]
		if strings.HasPrefix(f.Name(), ".") || !f.IsDir() {
			continue
		}
		matchDir := filepath.Join(directory, f.Name())
		for _, extension := range iconExtensions {
			testIcon := filepath.Join(matchDir, iconName+extension)
			if _, err := os.Stat(testIcon); err == nil {
				return testIcon
			}
		}
	}

	return ""
}

// lookupIconPathInTheme searches icon locations to find a match using a provided theme directory
func lookupIconPathInTheme(iconSize string, dir string, iconName string) string {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return ""
	}
	for _, extension := range iconExtensions {
		// Example is /usr/share/icons/icon_theme/32/apps/xterm.png
		testIcon := filepath.Join(dir, iconSize, "apps", iconName+extension)
		if _, err := os.Stat(testIcon); err == nil {
			return testIcon
		}
		// Example is /usr/share/icons/icon_theme/32x32/apps/xterm.png
		testIcon = filepath.Join(dir, iconSize+"x"+iconSize, "apps", iconName+extension)
		if _, err := os.Stat(testIcon); err == nil {
			return testIcon
		}

		// Example is /usr/share/icons/icon_theme/apps/32/xterm.png
		testIcon = filepath.Join(dir, "apps", iconSize, iconName+extension)
		if _, err := os.Stat(testIcon); err == nil {
			return testIcon
		}
		// Example is /usr/share/icons/icon_theme/apps/32x32/xterm.png
		testIcon = filepath.Join(dir, "apps", iconSize+"x"+iconSize, iconName+extension)
		if _, err := os.Stat(testIcon); err == nil {
			return testIcon
		}

		// Try this if the requested iconSize didn't exist
		// Example is /usr/share/icons/icon_theme/scalable/apps/xterm.png
		testIcon = filepath.Join(dir, "scalable", "apps", iconName+extension)
		if _, err := os.Stat(testIcon); err == nil {
			return testIcon
		}
		// Example is /usr/share/icons/icon_theme/apps/scalable/xterm.png
		testIcon = filepath.Join(dir, "apps", "scalable", iconName+extension)
		if _, err := os.Stat(testIcon); err == nil {
			return testIcon
		}
	}
	// If the requested icon wasn't found in the specific size or scalable dirs
	// of the theme try all sizes within theme and all icon dirs besides apps
	var subIconDirs = []string{"apps", "actions", "devices", "emblems", "mimetypes", "places", "status"}
	for _, joiner := range subIconDirs {
		testIcon := lookupAnyIconSizeInThemeDir(dir, joiner, iconName)
		if testIcon != "" {
			return testIcon
		}
	}
	return ""
}

//fdoLookupIconPath will take the name of an icon and find a matching image file
func fdoLookupIconPath(theme string, size int, iconName string) string {
	locationLookup := fdoLookupXdgDataDirs()
	iconTheme := theme
	iconSize := fmt.Sprintf("%d", size)
	for _, dataDir := range locationLookup {
		//Example is /usr/share/icons/icon_theme
		dir := filepath.Join(dataDir, "icons", iconTheme)
		iconPath := lookupIconPathInTheme(iconSize, dir, iconName)
		if iconPath != "" {
			return iconPath
		}
		indexTheme := filepath.Join(dir, "index.theme")
		if _, err := os.Stat(indexTheme); os.IsNotExist(err) {
			continue
		}
		file, err := os.Open(indexTheme)
		if err != nil {
			fyne.LogError("Could not open file", err)
			continue
		}
		defer file.Close()

		var inheritedThemes []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "Inherits=") {
				inherits := strings.SplitAfter(line, "=")
				inheritedThemes = strings.SplitN(inherits[1], ",", -1)
				break
			}
		}
		if len(inheritedThemes) == 0 {
			continue
		}
		for _, theme := range inheritedThemes {
			dir = filepath.Join(dataDir, "icons", theme)
			iconPath = lookupIconPathInTheme(iconSize, dir, iconName)
			if iconPath != "" {
				return iconPath
			}
		}
	}
	for _, dataDir := range locationLookup {
		//Hicolor is the default fallback theme - Example /usr/share/icons/icon_theme/hicolor
		dir := filepath.Join(dataDir, "icons", "hicolor")
		iconPath := lookupIconPathInTheme(iconSize, dir, iconName)
		if iconPath != "" {
			return iconPath
		}
	}
	for _, dataDir := range locationLookup {
		//Icons may be in the pixmaps directory - test before we do our final fallback
		for _, extension := range iconExtensions {
			iconPath := filepath.Join(dataDir, "pixmaps", iconName+extension)
			if _, err := os.Stat(iconPath); err == nil {
				return iconPath
			}
		}
	}
	for _, dataDir := range locationLookup {
		//Icon was not found in default theme or default fallback theme - Check all themes for a match
		//Example is /usr/share/icons
		files, err := ioutil.ReadDir(filepath.Join(dataDir, "icons"))
		if err != nil {
			continue
		}
		//Enter icon theme
		for _, f := range files {
			if strings.HasPrefix(f.Name(), ".") || !f.IsDir() {
				continue
			}

			//Example is /usr/share/icons/gnome
			iconPath := lookupIconPathInTheme(iconSize, filepath.Join(dataDir, "icons", f.Name()), iconName)
			if iconPath != "" {
				return iconPath
			}
		}
	}
	//No Icon Was Found
	return ""
}

func fdoLookupAvailableThemes() []string {
	var themes []string
	locationLookup := fdoLookupXdgDataDirs()
	for _, dataDir := range locationLookup {
		files, err := ioutil.ReadDir(filepath.Join(dataDir, "icons"))
		if err != nil {
			continue
		}
		//Enter icon theme
		for _, f := range files {
			if strings.HasPrefix(f.Name(), ".") || !f.IsDir() {
				continue
			}
			//Example is /usr/share/icons/gnome
			themes = append(themes, f.Name())
		}
	}
	return themes
}

//newFdoIconData creates and returns a struct that contains needed fields from a .desktop file
func newFdoIconData(desktopPath string) desktop.AppData {
	file, err := os.Open(desktopPath)
	if err != nil {
		fyne.LogError("Could not open file", err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	fdoApp := fdoApplicationData{name: "", iconName: "", iconPath: "", exec: ""}
	var currentSection string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "[") {
			currentSection = line
		}
		if currentSection != "[Desktop Entry]" {
			continue
		}
		if strings.HasPrefix(line, "Name=") {
			name := strings.SplitAfter(line, "=")
			fdoApp.name = name[1]
		} else if strings.HasPrefix(line, "Icon=") {
			icon := strings.SplitAfter(line, "=")
			fdoApp.iconName = icon[1]
			if _, err := os.Stat(icon[1]); err == nil {
				fdoApp.iconPath = icon[1]
			}
		} else if strings.HasPrefix(line, "Exec=") {
			exec := strings.SplitAfter(line, "=")
			fdoApp.exec = exec[1]
		}
	}
	if err := scanner.Err(); err != nil {
		fyne.LogError("Could not read file", err)
		return nil
	}
	return &fdoApp
}

type fdoIconProvider struct {
}

//AllApplications returns all of the available applications in a AppData slice
func (f *fdoIconProvider) AvailableApps() []desktop.AppData {
	return fdoLookupApplications()
}

//AvailableThemes returns all available icon themes in a string slice
func (f *fdoIconProvider) AvailableThemes() []string {
	return fdoLookupAvailableThemes()
}

//FindAppFromName matches an icon name to a location and returns an AppData interface
func (f *fdoIconProvider) FindAppFromName(appName string) desktop.AppData {
	return fdoLookupApplication(appName)
}

//FindIconFromPartialAppName returns a list of icons that match a partial name of an app and returns an AppData slice
func (f *fdoIconProvider) FindAppsMatching(appName string) []desktop.AppData {
	return fdoLookupApplicationsMatching(appName)
}

//FindAppFromWinInfo matches window information to an icon location and returns an AppData interface
func (f *fdoIconProvider) FindAppFromWinInfo(win desktop.Window) desktop.AppData {
	return fdoLookupApplicationWinInfo(win)
}

func findOneAppFromNames(f desktop.ApplicationProvider, names ...string) desktop.AppData {
	for _, name := range names {
		app := f.FindAppFromName(name)
		if app != nil {
			return app
		}
	}

	return nil
}

func appendAppIfExists(apps []desktop.AppData, app desktop.AppData) []desktop.AppData {
	if app == nil {
		return apps
	}

	return append(apps, app)
}

func (f *fdoIconProvider) DefaultApps() []desktop.AppData {
	var apps []desktop.AppData

	apps = appendAppIfExists(apps, findOneAppFromNames(f, "xfce4-terminal", "gnome-terminal", "xterm"))
	apps = appendAppIfExists(apps, findOneAppFromNames(f, "chromium", "google-chrome", "firefox"))
	apps = appendAppIfExists(apps, findOneAppFromNames(f, "sylpheed", "thunderbird", "evolution"))
	apps = appendAppIfExists(apps, f.FindAppFromName("gimp"))

	return apps
}

// NewFDOIconProvider returns a new icon provider following the FreeDesktop.org specifications
func NewFDOIconProvider() desktop.ApplicationProvider {
	return &fdoIconProvider{}
}
