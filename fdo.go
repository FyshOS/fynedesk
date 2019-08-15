package desktop

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

//IconData is a structure that contains information about .desktop files
type IconData struct {
	Name     string // Application name
	IconName string // Icon name
	IconPath string // Icon path
	Exec     string // Command to execute application
}

//FdoLookupXdgDataDirs returns a string slice of all XDG_DATA_DIRS
func FdoLookupXdgDataDirs() []string {
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

//FdoLookupApplication looks up an application by name and returns an IconData struct
func FdoLookupApplication(appName string) *IconData {
	locationLookup := FdoLookupXdgDataDirs()
	for _, dataDir := range locationLookup {
		testLocation := filepath.Join(dataDir, "applications", appName+".desktop")
		if _, err := os.Stat(testLocation); err == nil {
			return NewIconData(testLocation)
		}
	}
	return nil
}

//FdoLookupApplicationWinInfo looks up an application based on window info
func FdoLookupApplicationWinInfo(title string, classes []string, command string, iconName string) *IconData {
	desktop := FdoLookupApplication(title)
	if desktop != nil {
		return desktop
	}
	for _, class := range classes {
		desktop := FdoLookupApplication(class)
		if desktop != nil {
			return desktop
		}
	}
	desktop = FdoLookupApplication(command)
	if desktop != nil {
		return desktop
	}
	desktop = FdoLookupApplication(iconName)
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
func lookupIconPathInTheme(dir string, iconName string) string {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return ""
	}
	size := fmt.Sprintf("%d", iconSize)
	for _, extension := range iconExtensions {
		//Example is /usr/share/icons/icon_theme/32x32/apps/xterm.png
		testIcon := filepath.Join(dir, size, "apps", iconName+extension)
		if _, err := os.Stat(testIcon); err == nil {
			return testIcon
		}
		testIcon = filepath.Join(dir, size+"x"+size, "apps", iconName+extension)
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
		//Example is /usr/share/icons/icon_theme/scalable/apps/xterm.png - Try this if the requested size didn't exist
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

//FdoLookupIconPath will take the name of an icon and find a matching image file
func FdoLookupIconPath(iconName string) string {
	locationLookup := FdoLookupXdgDataDirs()
	for _, dataDir := range locationLookup {
		//Example is /usr/share/icons/icon_theme
		dir := filepath.Join(dataDir, "icons", iconTheme)
		iconPath := lookupIconPathInTheme(dir, iconName)
		if iconPath != "" {
			return iconPath
		}
	}
	for _, dataDir := range locationLookup {
		//Hicolor is the default fallback theme - Example /usr/share/icons/icon_theme/hicolor
		dir := filepath.Join(dataDir, "icons", "hicolor")
		iconPath := lookupIconPathInTheme(dir, iconName)
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
				lookupIconPathInTheme(filepath.Join(dataDir, "icons", f.Name()), iconName)
			}
		}
	}
	//No Icon Was Found
	return ""
}

//NewIconData creates and returns a struct that contains needed fields from a .desktop file
func NewIconData(desktopPath string) *IconData {
	file, err := os.Open(desktopPath)
	if err != nil {
		fyne.LogError("Could not open file", err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	IconData := IconData{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Name=") {
			name := strings.SplitAfter(line, "=")
			IconData.Name = name[1]
		} else if strings.HasPrefix(line, "Icon=") {
			icon := strings.SplitAfter(line, "=")
			IconData.IconName = icon[1]
			IconData.IconPath = FdoLookupIconPath(IconData.IconName)
		} else if strings.HasPrefix(line, "Exec=") {
			exec := strings.SplitAfter(line, "=")
			IconData.Exec = exec[1]
		} else {
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		fyne.LogError("Could not read file", err)
		return nil
	}
	return &IconData
}
