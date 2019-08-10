package desktop

import (
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
)

var fyconSize = 32
var fyconTheme = "Papirus"

func formatVariable(name string) string {
	str := strings.Replace(name, "-", "", -1)
	return strings.Replace(str, "_", "", -1)
}

func newBar() fyne.CanvasObject {
	bar := NewAppBar()
	bar.Append(layout.NewSpacer())

	icons := []string{"xterm", "gimp", "google-chrome", "firefox", "xterm", "gimp", "google-chrome", "firefox"}
	for _, icon := range icons {
		fdoDesktop := FdoLookupApplication(icon)
		if fdoDesktop != nil {
			if fdoDesktop.IconPath != "" {
				bytes, err := ioutil.ReadFile(fdoDesktop.IconPath)
				if err != nil {
					log.Print(err)
					continue
				}
				res := fyne.NewStaticResource(formatVariable(strings.ToLower(filepath.Base(fdoDesktop.IconPath))), bytes)
				fycon := NewFycon(res)
				fycon.OnTapped = func() {
					exec.Command(fdoDesktop.Exec).Start()
				}
				bar.Append(fycon)
			}
		}
	}

	bar.Append(layout.NewSpacer())

	return bar
}
