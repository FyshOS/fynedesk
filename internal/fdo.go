package internal

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne"
)

var iconExtensions = []string{".png", ".svg"}

//FdoIconData is a structure that contains information about .desktop files
type FdoIconData struct {
	name     string // Application name
	iconName string // Icon name
	iconPath string // Icon path
	exec     string // Command to execute application
}

//Name returns the name associated with an fdo app
func (data *FdoIconData) Name() string {
	return data.name
}

//IconName returns the name of the icon that an fdo app wishes to use
func (data *FdoIconData) IconName() string {
	return data.iconName
}

//IconPath returns the path of the icon that an fdo app wishes to use
func (data *FdoIconData) IconPath() string {
	return data.iconPath
}

//Exec returns the command used to execute the fdo app
func (data *FdoIconData) Exec() string {
	return data.exec
}

//fdoLookupXdgDataDirs returns a string slice of all XDG_DATA_DIRS
func fdoLookupXdgDataDirs() []string {
	dataLocation := os.Getenv("XDG_DATA_DIRS")
	locationLookup := strings.Split(dataLocation, ":")
	if len(locationLookup) == 1 && locationLookup[0] == dataLocation {
		var fallbackLocations []string
		fallbackLocations = append(fallbackLocations, "/usr/local/share")
		fallbackLocations = append(fallbackLocations, "/usr/share")
		return fallbackLocations
	}
	return locationLookup
}

//FdoLookupApplication looks up an application by name and returns an FdoIconData struct
func FdoLookupApplication(theme string, size int, appName string) *FdoIconData {
	locationLookup := fdoLookupXdgDataDirs()
	for _, dataDir := range locationLookup {
		testLocation := filepath.Join(dataDir, "applications", appName+".desktop")
		if _, err := os.Stat(testLocation); err == nil {
			return newFdoIconData(theme, size, testLocation)
		}
	}
	return nil
}

//FdoLookupApplicationWinInfo looks up an application based on window info and returns an FdoIconData struct
func FdoLookupApplicationWinInfo(theme string, size int, title string, classes []string, command string, iconName string) *FdoIconData {
	desktop := FdoLookupApplication(theme, size, title)
	if desktop != nil {
		return desktop
	}
	for _, class := range classes {
		desktop := FdoLookupApplication(theme, size, class)
		if desktop != nil {
			return desktop
		}
	}
	desktop = FdoLookupApplication(theme, size, command)
	if desktop != nil {
		return desktop
	}
	desktop = FdoLookupApplication(theme, size, iconName)
	return desktop
}

//reverseWalkDirectoryMatch walks file inside of a directory in reverse order to make sure larger sized icons are found first
func reverseWalkDirectoryMatch(directory string, joiner string, iconName string) string {
	//If the requested icon wasn't found in the Example 32x32 or scalable dirs of the theme, try other sizes within theme
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		fyne.LogError("Could not read directory", err)
		return ""
	}
	//Lets walk the files in reverse so bigger icons are selected first (Unless it is a 3 digit icon size like 128)
	for i := len(files) - 1; i >= 0; i-- {
		f := files[i]
		if strings.HasPrefix(f.Name(), ".") == true || f.IsDir() == true {
			continue
		}
		matchDir := filepath.Join(directory, f.Name())
		for _, extension := range iconExtensions {
			//Example is /usr/share/icons/icon_theme/64x64/apps/xterm.png
			var testIcon = ""
			if joiner != "" {
				testIcon = filepath.Join(matchDir, joiner, iconName+extension)
			} else {
				testIcon = filepath.Join(matchDir, iconName+extension)
			}
			if _, err := os.Stat(testIcon); err == nil {
				return testIcon
			}
		}
	}
	return ""
}

//lookupIconPathInTheme searches icon locations to find a match using a provided theme directory
func lookupIconPathInTheme(iconSize string, dir string, iconName string) string {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return ""
	}
	for _, extension := range iconExtensions {
		//Example is /usr/share/icons/icon_theme/32x32/apps/xterm.png
		testIcon := filepath.Join(dir, iconSize, "apps", iconName+extension)
		if _, err := os.Stat(testIcon); err == nil {
			return testIcon
		}
		testIcon = filepath.Join(dir, iconSize+"x"+iconSize, "apps", iconName+extension)
		if _, err := os.Stat(testIcon); err == nil {
			return testIcon
		}
		if _, err := os.Stat(testIcon); err == nil {
			return testIcon
		}
		if _, err := os.Stat(testIcon); err == nil {
			return testIcon
		}
	}
	for _, extension := range iconExtensions {
		//Example is /usr/share/icons/icon_theme/scalable/apps/xterm.png - Try this if the requested iconSize didn't exist
		testIcon := filepath.Join(dir, "scalable", "apps", iconName+extension)
		if _, err := os.Stat(testIcon); err == nil {
			return testIcon
		}
		testIcon = filepath.Join(dir, "apps", "scalable", iconName+extension)
		if _, err := os.Stat(testIcon); err == nil {
			return testIcon
		}
	}
	//If the requested icon wasn't found in the Example 32x32 or scalable dirs of the theme, try other sizes within theme
	testIcon := reverseWalkDirectoryMatch(dir, "apps", iconName)
	if testIcon != "" {
		return testIcon
	}

	//One last chance at finding it - assume apps comes before size in path name
	testIcon = reverseWalkDirectoryMatch(filepath.Join(dir, "apps"), "", iconName)

	return testIcon
}

//fdoLookupIconPath will take the name of an icon and find a matching image file
func fdoLookupIconPath(theme string, size int, iconName string) string {
	locationLookup := fdoLookupXdgDataDirs()
	iconTheme := theme
	iconSize := fmt.Sprintf("%d", size)
	for _, dataDir := range locationLookup {
		//Example is /usr/share/icons/icon_theme
		dir := filepath.Join(dataDir, "icons", iconTheme)
		iconPath := lookupIconPathInTheme(iconSize, dir, iconTheme)
		if iconPath != "" {
			return iconPath
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
		//Icon was not found in default theme or default fallback theme - Check all themes for a match
		//Example is /usr/share/icons
		files, err := ioutil.ReadDir(dataDir + "icons")
		if err != nil {
			fyne.LogError("Could not read directory", err)
			continue
		}
		//Enter icon theme
		for _, f := range files {
			if strings.HasPrefix(f.Name(), ".") == true || f.IsDir() == true {
				continue
			}
			if f.IsDir() {
				//Example is /usr/share/icons/gnome
				lookupIconPathInTheme(iconSize, filepath.Join(dataDir, "icons", f.Name()), iconName)
			}
		}
	}
	//No Icon Was Found
	return ""
}

//newFdoIconData creates and returns a struct that contains needed fields from a .desktop file
func newFdoIconData(theme string, size int, desktopPath string) *FdoIconData {
	file, err := os.Open(desktopPath)
	if err != nil {
		fyne.LogError("Could not open file", err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	fdoApp := FdoIconData{name: "", iconName: "", iconPath: "", exec: ""}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Name=") {
			name := strings.SplitAfter(line, "=")
			fdoApp.name = name[1]
		} else if strings.HasPrefix(line, "Icon=") {
			icon := strings.SplitAfter(line, "=")
			fdoApp.iconName = icon[1]
			fdoApp.iconPath = fdoLookupIconPath(theme, size, fdoApp.iconName)
		} else if strings.HasPrefix(line, "Exec=") {
			exec := strings.SplitAfter(line, "=")
			fdoApp.exec = exec[1]
		} else {
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		fyne.LogError("Could not read file", err)
		return nil
	}
	return &fdoApp
}
