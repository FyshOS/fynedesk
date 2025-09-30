// Note that you need to have github.com/knightpp/dbus-codegen-go installed
//go:generate dbus-codegen-go -prefix org.freedesktop -package screensaver -output generated/screensaver.go dbus/ScreenSaver.xml

package ui

import (
	"math/rand"
	"os/exec"
	"time"

	screensaver "fyshos.com/fynedesk/internal/ui/generated"
	"github.com/FyshOS/saver"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"

	"fyne.io/fyne/v2"
)

var (
	inhibitCount = 0
	lastActivity = time.Now()
)

func (l *desktop) startXscreensaver() {
	_, err := exec.LookPath("xscreensaver")
	if err != nil {
		fyne.LogError("xscreensaver command not found", err)
		return
	}
	err = exec.Command("xscreensaver", "-no-splash").Start()
	if err != nil {
		fyne.LogError("Failed to lock screen", err)
	}
}

func (l *desktop) TriggerScreenSaver() {
	s := saver.NewScreenSaver(nil)
	s.ClockFormat = l.settings.ClockFormatting()
	if l.settings.ScreenSaverClock() {
		s.Label = "(clock)"
	} else {
		s.Label = l.settings.ScreenSaverLabel()
	}
	s.Lock = true

	l.wm.ShowScreensaver(s)
}

func (l *desktop) DelayScreenSaver() {
	lastActivity = time.Now()
}

func (l *desktop) watchScreenActivity() {
	watchDBus()
	idle := false
	to := time.NewTicker(5 * time.Second)

	for range to.C {
		if inhibitCount == 0 && lastActivity.Add(time.Minute*5).Before(time.Now()) {

			if !idle {
				idle = true

				fyne.Do(l.TriggerScreenSaver)
			}
		} else {
			idle = false
		}
	}
}

func watchDBus() {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		fyne.LogError("failed to connect to DBus to watch for screensaver inhibits", err)
		return
	}

	name := "org.freedesktop.ScreenSaver"
	r, err := conn.RequestName(name, dbus.NameFlagDoNotQueue)
	if err != nil || r != dbus.RequestNameReplyPrimaryOwner {
		fyne.LogError("could not watch DBus screensaver, another is registered", err)
		return
	}

	s := &screenSaverWatcher{}
	path := "/org/freedesktop/ScreenSaver"
	err = conn.ExportAll(s, dbus.ObjectPath(path), "org.freedesktop.ScreenSaver")
	if err != nil {
		fyne.LogError("failed to export inhibits", err)
		return
	}

	node := introspect.Node{
		Name: path,
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			prop.IntrospectData,
			screensaver.IntrospectDataScreenSaver,
		},
	}
	err = conn.Export(introspect.NewIntrospectable(&node), dbus.ObjectPath(path),
		"org.freedesktop.DBus.Introspectable")
	if err != nil {
		fyne.LogError("could not export our node data", err)
	}
}

type screenSaverWatcher struct {
}

func (s *screenSaverWatcher) Inhibit(_ dbus.Sender, who, why string) (uint, *dbus.Error) {
	id := rand.Uint32()
	inhibitCount++

	// TODO also check these are still alive every so often
	return uint(id), nil
}

func (s *screenSaverWatcher) UnInhibit(_ dbus.Sender, cookie uint32) *dbus.Error {
	// TODO compare to the cookies logged
	inhibitCount--
	return nil
}
