package desktop

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

var fyconSize = 32
var fyconTheme = "Papirus"

func locateApplication(appName string) *FybarDesktop {
	dataLocation := os.Getenv("XDG_DATA_DIRS")
	locationLookup := strings.Split(dataLocation, ":")
	for _, dataDir := range locationLookup {
		testLocation := dataDir + "/applications/" + appName + ".desktop"
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
				res := fyne.NewStaticResource(formatVariable(strings.ToLower(fybarDesktop.IconName+".svg")), bytes)
				fycon := widget.NewIcon(res)
				//Uncomment to use Fycon which is a clickable icon.  You have to switch themes to get them to show, and their sizing is wonky compared to widget.Icon
				//fycon := NewFycon(res)
				//fycon.OnTapped = func() {
				//	exec.Command(fybarDesktop.Exec).Start()
				//}
				fybar.Append(fycon)
			}
		}
	}

	fybar.Append(layout.NewSpacer())

	return fybar
}
