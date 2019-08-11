package desktop

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne"
)

var iconExtensions = []string{".png", ".svg", ".xpm"}

//FdoDesktop contains the information from Linux .desktop files
type FdoDesktop struct {
	Name     string
	IconName string
	IconPath string
	Exec     string
}

//FdoResourceFormat turns a string into a usable name for fyne.Resources
func FdoResourceFormat(name string) string {
	str := strings.Replace(name, "-", "", -1)
	return strings.Replace(str, "_", "", -1)
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

//FdoLookupApplication looks up an application by name and returns a FdoDesktop
func FdoLookupApplication(appName string) *FdoDesktop {
	locationLookup := FdoLookupXdgDataDirs()
	for _, dataDir := range locationLookup {
		testLocation := filepath.Join(dataDir, "applications", appName+".desktop")
		if _, err := os.Stat(testLocation); os.IsNotExist(err) {
			continue
		} else {
			return NewFdoDesktop(testLocation)
		}
	}
	return nil
}

//FdoLookupApplicationWinInfo looks up an application based on window info
func FdoLookupApplicationWinInfo(title string, classes []string, command string, iconName string) *FdoDesktop {
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

func reverseWalkDirectoryMatch(directory string, joiner string, iconName string) string {
	//If the requested icon wasn't found in the Example 32x32 or scalable dirs of the theme, try other sizes within theme
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		fyne.LogError("", err)
		return ""
	}
	//Lets walk the files in reverse so bigger icons are selected first (Unless it is a 3 digit icon size like 128)
	for i := len(files) - 1; i >= 0; i-- {
		f := files[i]
		if strings.HasPrefix(f.Name(), ".") == false {
			if f.IsDir() {
				matchDir := filepath.Join(directory, f.Name())
				for _, extension := range iconExtensions {
					//Example is /usr/share/icons/icon_theme/64x64/apps/xterm.png
					var testIcon = ""
					if joiner != "" {
						testIcon = filepath.Join(matchDir, joiner, iconName+extension)
					} else {
						testIcon = filepath.Join(matchDir, iconName+extension)
					}
					if _, err := os.Stat(testIcon); os.IsNotExist(err) {
						continue
					} else {
						return testIcon
					}
				}
			}
		}
	}
	return ""
}

func lookupIconPathInTheme(dir string, iconName string) string {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return ""
	}
	for _, extension := range iconExtensions {
		//Example is /usr/share/icons/icon_theme/32x32/apps/xterm.png
		testIcon := filepath.Join(dir, string(fyconSize), "apps", iconName+extension)
		if _, err := os.Stat(testIcon); os.IsNotExist(err) {
			testIcon := filepath.Join(dir, string(fyconSize)+"x"+string(fyconSize), "apps", iconName+extension)
			if _, err := os.Stat(testIcon); os.IsNotExist(err) {
				testIcon := filepath.Join(dir, "apps", string(fyconSize), iconName+extension)
				if _, err := os.Stat(testIcon); os.IsNotExist(err) {
					testIcon := filepath.Join(dir, "apps", string(fyconSize)+"x"+string(fyconSize), iconName+extension)
					if _, err := os.Stat(testIcon); os.IsNotExist(err) {
						continue
					} else {
						return testIcon
					}
				} else {
					return testIcon
				}
			} else {
				return testIcon
			}
		} else {
			return testIcon
		}
	}
	for _, extension := range iconExtensions {
		//Example is /usr/share/icons/icon_theme/scalable/apps/xterm.png - Try this if the requested size didn't exist
		testIcon := filepath.Join(dir, "scalable", "apps", iconName+extension)
		if _, err := os.Stat(testIcon); os.IsNotExist(err) {
			testIcon := filepath.Join(dir, "apps", "scalable", iconName+extension)
			if _, err := os.Stat(testIcon); os.IsNotExist(err) {
				continue
			} else {
				return testIcon
			}
		} else {
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

//FdoLookupIconPath will take the name of an Icon and find a matching image file
func FdoLookupIconPath(iconName string) string {
	locationLookup := FdoLookupXdgDataDirs()
	for _, dataDir := range locationLookup {
		//Example is /usr/share/icons/icon_theme
		dir := filepath.Join(dataDir, "icons", fyconTheme)
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
			fyne.LogError("", err)
			continue
		}
		//Enter icon theme
		for _, f := range files {
			if strings.HasPrefix(f.Name(), ".") == false {
				if f.IsDir() {
					//Example is /usr/share/icons/gnome

					subFiles, err := ioutil.ReadDir(filepath.Join(dataDir, "icons", f.Name()))
					if err != nil {
						fyne.LogError("", err)
						continue
					}
					//Let's walk the files in reverse order so larger sizes come first - Except 3 digit icon size like 128
					for i := len(subFiles) - 1; i >= 0; i-- {
						subf := subFiles[i]
						if strings.HasPrefix(subf.Name(), ".") == false {
							if subf.IsDir() {
								//Example is /usr/share/icons/gnome/32x32
								fallbackDir := filepath.Join(dataDir, "icons", f.Name(), subf.Name())
								extensions := []string{".png", ".svg", ".xpm"}
								for _, extension := range extensions {
									//Example is /usr/share/icons/gnome/32x32/apps/xterm.png
									testIcon := filepath.Join(fallbackDir, "apps", iconName+extension)
									if _, err := os.Stat(testIcon); os.IsNotExist(err) {
										if subf.Name() == "apps" {
											reverseWalkDirectoryMatch(fallbackDir, "", iconName)
										}
										continue
									} else {
										return testIcon
									}
								}
							}
						}
					}
				}
			}
		}
	}
	//No Icon Was Found
	return ""
}

//NewFdoDesktop returns a new instance of FdoDesktop
func NewFdoDesktop(desktopPath string) *FdoDesktop {
	file, err := os.Open(desktopPath)
	if err != nil {
		fyne.LogError("", err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	fdoDesktop := FdoDesktop{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Name=") {
			name := strings.SplitAfter(line, "=")
			fdoDesktop.Name = name[1]
		} else if strings.HasPrefix(line, "Icon=") {
			icon := strings.SplitAfter(line, "=")
			fdoDesktop.IconName = icon[1]
			fdoDesktop.IconPath = FdoLookupIconPath(fdoDesktop.IconName)
		} else if strings.HasPrefix(line, "Exec=") {
			exec := strings.SplitAfter(line, "=")
			fdoDesktop.Exec = exec[1]
		} else {
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		fyne.LogError("", err)
		return nil
	}
	return &fdoDesktop
}
