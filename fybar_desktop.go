package desktop

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

//FybarDesktop contains the information from Linux .desktop files
type FybarDesktop struct {
	Name     string
	IconName string
	IconPath string
	Exec     string
}

func lookupIconPathInTheme(dir string, iconName string) string {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return ""
	}
	extensions := []string{".png", ".svg", ".xpm"}
	for _, extension := range extensions {
		//Example is /usr/share/icons/icon_theme/32x32/apps/xterm.png
		testIcon := filepath.Join(dir, string(fyconSize)+"x"+string(fyconSize), "apps", iconName+extension)
		if _, err := os.Stat(testIcon); os.IsNotExist(err) {
			continue
		} else {
			return testIcon
		}
	}
	for _, extension := range extensions {
		//Example is /usr/share/icons/icon_theme/scalable/apps/xterm.png - Try this if the requested size didn't exist
		testIcon := filepath.Join(dir, "scalable", "apps", iconName+extension)
		if _, err := os.Stat(testIcon); os.IsNotExist(err) {
			continue
		} else {
			return testIcon
		}
	}
	//If the requested icon wasn't found in the Example 32x32 or scalable dirs of the theme, try other sizes within theme
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Print(err)
		return ""
	}
	//Lets walk the files in reverse so bigger icons are selected first (Unless it is a 3 digit icon size like 128)
	for i := len(files) - 1; i >= 0; i-- {
		f := files[i]
		if strings.HasPrefix(f.Name(), ".") == false {
			if f.IsDir() {
				fallbackDir := filepath.Join(dir, f.Name())
				for _, extension := range extensions {
					//Example is /usr/share/icons/icon_theme/64x64/apps/xterm.png
					testIcon := filepath.Join(fallbackDir, "/apps/", iconName+extension)
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

//LookupIconPath will take the name of an Icon and find a matching image file
func LookupIconPath(iconName string) string {
	locationLookup := lookupXDGdirs()
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
			log.Print(err)
			continue
		}
		//Enter icon theme
		for _, f := range files {
			if strings.HasPrefix(f.Name(), ".") == false {
				if f.IsDir() {
					//Example is /usr/share/icons/gnome
					subFiles, err := ioutil.ReadDir(filepath.Join(dataDir, "icons", f.Name()))
					if err != nil {
						log.Print(err)
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

//NewFybarDesktop returns a new instance of FybarDesktop
func NewFybarDesktop(desktopPath string) *FybarDesktop {
	file, err := os.Open(desktopPath)
	if err != nil {
		log.Print(err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	fbIcon := FybarDesktop{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Name=") {
			name := strings.SplitAfter(line, "=")
			fbIcon.Name = name[1]
		} else if strings.HasPrefix(line, "Icon=") {
			icon := strings.SplitAfter(line, "=")
			fbIcon.IconName = icon[1]
			fbIcon.IconPath = LookupIconPath(fbIcon.IconName)
		} else if strings.HasPrefix(line, "Exec=") {
			exec := strings.SplitAfter(line, "=")
			fbIcon.Exec = exec[1]
		} else {
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		log.Print(err)
		return nil
	}
	return &fbIcon
}
