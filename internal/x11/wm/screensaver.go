//go:build linux || openbsd || freebsd || netbsd
// +build linux openbsd freebsd netbsd

package wm

import (
	"log"
	"os/exec"
	"time"

	"fyshos.com/fynedesk"
	"github.com/BurntSushi/xgb/screensaver"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/FyshOS/saver"

	"fyne.io/fyne/v2"
)

func (x *x11WM) initScreensaver() {
	err := screensaver.Init(x.x.Conn())
	if err != nil {
		log.Println("Failed to init screensaver extension")
		return
	}

	//screensaver.SelectInput(conn.Conn(), xproto.Drawable(conn.Screen().Root),
	//	screensaver.EventNotifyMask)
	go x.watchScreensaver()
}

func (x *x11WM) watchScreensaver() {
	// more complex than time.Timer so that when the OS sleeps it does not stack ticks...
	wait := make(chan struct{})
	time.AfterFunc(time.Second, func() {
		wait <- struct{}{}
	})
	previous := time.Now()

	for range wait {
		info, err := screensaver.QueryInfo(x.x.Conn(), xproto.Drawable(x.x.Screen().Root)).Reply()
		if err != nil {
			log.Println("ERR", err)
			continue
		}

		now := time.Now()
		slept := now.Sub(previous).Seconds() > 2 // skipped a tick
		previous = now
		if slept {
			fynedesk.Instance().TriggerScreenSaver()
		} else if info.MsSinceUserInput <= 1500 {
			fynedesk.Instance().DelayScreenSaver()
		}

		time.AfterFunc(time.Second, func() {
			wait <- struct{}{}
		})
	}
}

var screenSaverActive bool

func (x *x11WM) ShowScreensaver(s *saver.ScreenSaver) {
	if fynedesk.Instance().Settings().ScreenSaverType() == "XScreensaver" {
		task := "-activate"
		if s.Lock {
			task = "-lock"
		}
		cmd := exec.Command("xscreensaver-command", task)
		cmd.Start()
		return
	}

	if screenSaverActive {
		return
	}

	screenSaverActive = true
	s.OnUnlocked = func() {
		screenSaverActive = false
	}

	go func() {
		time.Sleep(time.Millisecond * 100)
		fyne.Do(s.ShowWindows)
	}()
}
