package desktop

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
)

var appBar *Bar
var fyconSize = 32
var fyconTheme = "Papirus"

type barStackListener struct {
}

func barCreateFycon(taskbar bool, fdoDesktop *FdoDesktop) *Fycon {
	if fdoDesktop != nil {
		if fdoDesktop.IconPath != "" {
			bytes, err := ioutil.ReadFile(fdoDesktop.IconPath)
			if err != nil {
				fyne.LogError("", err)
				return nil
			}
			res := fyne.NewStaticResource(FdoResourceFormat(strings.ToLower(filepath.Base(fdoDesktop.IconPath))), bytes)
			fycon := NewFycon(res)
			if taskbar == false {
				fycon.OnTapped = func() {
					command := strings.Split(fdoDesktop.Exec, " ")
					if len(command) > 1 {
						args := fmt.Sprintf("%s", command[1:])
						exec.Command(command[0], args).Start()
					} else {
						exec.Command(command[0]).Start()
					}
				}
			}
			return fycon
		}
	}
	return nil
}

func (bsl *barStackListener) WindowAdded(win Window) {
	fdoDesktop := FdoLookupApplicationWinInfo(win.Title(), win.Class(), win.Command(), win.IconName())
	fycon := barCreateFycon(true, fdoDesktop)
	if fycon != nil {
		fycon.OnTapped = func() {
			win.Focus()
		}
		appBar.AppendTaskbar(fycon)
	}
}

func (bsl *barStackListener) WindowRemoved(win Window) {
}

func newBar(wm WindowManager) fyne.CanvasObject {
	appBar = NewAppBar()
	appBar.Append(layout.NewSpacer())

	if wm != nil {
		bsl := &barStackListener{}
		wm.AddStackListener(bsl)
	}
	icons := []string{"xterm", "gimp", "google-chrome", "firefox", "xterm", "gimp", "google-chrome", "firefox"}
	for _, icon := range icons {
		fdoDesktop := FdoLookupApplication(icon)
		fycon := barCreateFycon(false, fdoDesktop)
		if fycon != nil {
			appBar.Append(barCreateFycon(false, fdoDesktop))
		}
	}
	appBar.AppendSeparator()
	appBar.Append(layout.NewSpacer())

	return appBar
}
