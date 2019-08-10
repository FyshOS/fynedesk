package desktop

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
)

var fyconSize = 32
var fyconTheme = "Papirus"

func lookupXDGdirs() []string {
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

func locateApplication(appName string) *FybarDesktop {
	locationLookup := lookupXDGdirs()
	for _, dataDir := range locationLookup {
		testLocation := filepath.Join(dataDir, "applications", appName+".desktop")
		if _, err := os.Stat(testLocation); os.IsNotExist(err) {
			continue
		} else {
			return NewFybarDesktop(testLocation)
		}
	}
	return nil
}

func formatVariable(name string) string {
	str := strings.Replace(name, "-", "", -1)
	return strings.Replace(str, "_", "", -1)
}

func newBar() fyne.CanvasObject {
	fybar := NewHFybar()
	fybar.Append(layout.NewSpacer())

	icons := []string{"xterm", "gimp", "google-chrome", "firefox", "xterm", "gimp", "google-chrome", "firefox"}
	for _, icon := range icons {
		fybarDesktop := locateApplication(icon)
		if fybarDesktop != nil {
			if fybarDesktop.IconPath != "" {
				bytes, err := ioutil.ReadFile(fybarDesktop.IconPath)
				if err != nil {
					log.Print(err)
					continue
				}
				res := fyne.NewStaticResource(formatVariable(strings.ToLower(filepath.Base(fybarDesktop.IconPath))), bytes)
				fycon := NewFycon(res)
				fycon.OnTapped = func() {
					exec.Command(fybarDesktop.Exec).Start()
				}
				fybar.Append(fycon)
			}
		}
	}

	fybar.Append(layout.NewSpacer())

	return fybar
}
