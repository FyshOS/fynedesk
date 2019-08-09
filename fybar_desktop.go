package desktop

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

//FybarDesktop contains the information from Linux .desktop files
type FybarDesktop struct {
	Name     string
	IconName string
	IconPath string
	Exec     string
}

//LookupIconPath will take the name of an Icon and find a matching image file
func LookupIconPath(iconName string) string {
	var iconPath = ""
	dataLocation := os.Getenv("XDG_DATA_DIRS")
	locationLookup := strings.Split(dataLocation, ":")
	for _, dataDir := range locationLookup {
		testLocation := dataDir + "/icons/" + fyconTheme
		if _, err := os.Stat(testLocation); os.IsNotExist(err) {
			continue
		} else {
			dir := dataDir + "/icons/" + fyconTheme
			extensions := []string{".png", ".svg", ".xpm"}
			for _, extension := range extensions {
				testIcon := dir + "/" + string(fyconSize) + "x" + string(fyconSize) + "/apps/" + iconName + extension
				if _, err := os.Stat(testIcon); os.IsNotExist(err) {
					continue
				} else {
					iconPath = testIcon
					return iconPath
				}
			}
			for _, extension := range extensions {
				testIcon := dir + "/scalable/apps/" + iconName + extension
				if _, err := os.Stat(testIcon); os.IsNotExist(err) {
					continue
				} else {
					iconPath = testIcon
					return iconPath
				}
			}
			files, err := ioutil.ReadDir(dir)
			if err != nil {
				log.Print(err)
				continue
			}
			for _, f := range files {
				if strings.HasPrefix(f.Name(), ".") == false {
					if f.IsDir() {
						fallbackDir := dir + "/" + f.Name()
						for _, extension := range extensions {
							testIcon := fallbackDir + "/apps/" + iconName + extension
							if _, err := os.Stat(testIcon); os.IsNotExist(err) {
								continue
							} else {
								iconPath = testIcon
								return iconPath
							}
						}
					}
				}
			}
			files, err = ioutil.ReadDir(dataDir + "/icons")
			if err != nil {
				log.Print(err)
				continue
			}
			for _, f := range files {
				if strings.HasPrefix(f.Name(), ".") == false {
					if f.IsDir() {
						subFiles, err := ioutil.ReadDir(dataDir + "/icons/" + f.Name())
						if err != nil {
							log.Print(err)
							continue
						}
						for _, subf := range subFiles {
							if strings.HasPrefix(subf.Name(), ".") == false {
								if subf.IsDir() {
									fallbackDir := dataDir + "/icons/" + f.Name() + "/" + subf.Name()
									for _, extension := range extensions {
										testIcon := fallbackDir + "/apps/" + iconName + extension
										if _, err := os.Stat(testIcon); os.IsNotExist(err) {
											continue
										} else {
											iconPath = testIcon
											return iconPath
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
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
