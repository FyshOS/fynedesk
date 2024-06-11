//go:build linux || openbsd || freebsd || netbsd
// +build linux openbsd freebsd netbsd

package wm

import (
	"time"

	"github.com/BurntSushi/xgb/xproto"

	"fyshos.com/fynedesk"
	"fyshos.com/fynedesk/internal/ui"
)

var (
	// switcherInstance and the methods below manage the X to ui.Switcher bindings.
	// This is needed due to the way that releasing Super is only reported if we grab the whole keyboard.
	// Therefore the UI cannot handle keyboard input - so we add it here instead.
	switcherInstance *ui.Switcher

	// ignoreSwitcher helps us track where key pres+lift is faster than window show
	ignoreSwitcher bool
)

func (x *x11WM) applyAppSwitcher() {
	if switcherInstance == nil {
		ignoreSwitcher = true
		go func() {
			time.Sleep(time.Second / 4)
			ignoreSwitcher = false
		}()
	} else {
		go switcherInstance.HideApply()
	}

	xproto.UngrabKeyboard(x.x.Conn(), xproto.TimeCurrentTime)
	windowClientListStackingUpdate(x)
	switcherInstance = nil
}

func (x *x11WM) cancelAppSwitcher() {
	if switcherInstance == nil {
		ignoreSwitcher = true
		go func() {
			time.Sleep(time.Second / 4)
			ignoreSwitcher = false
		}()
	} else {
		go switcherInstance.HideCancel()
	}

	xproto.UngrabKeyboard(x.x.Conn(), xproto.TimeCurrentTime)
	switcherInstance = nil
}

func (x *x11WM) nextAppSwitcher() {
	if switcherInstance == nil {
		return
	}

	go switcherInstance.Next()
}

func (x *x11WM) previousAppSwitcher() {
	if switcherInstance == nil {
		return
	}

	go switcherInstance.Previous()
}

func (x *x11WM) showOrSelectAppSwitcher(reverse bool) {
	var visible []fynedesk.Window
	for _, win := range x.clients {
		if win.Desktop() == fynedesk.Instance().Desktop() || win.Pinned() {
			visible = append(visible, win)
		}
	}
	if len(visible) <= 1 {
		return
	}
	xproto.GrabKeyboard(x.x.Conn(), true, x.x.RootWin(), xproto.TimeCurrentTime, xproto.GrabModeAsync, xproto.GrabModeAsync)

	if switcherInstance != nil {
		if reverse {
			switcherInstance.Previous()
		} else {
			switcherInstance.Next()
		}

		return
	}

	go func() {
		if reverse {
			win := ui.NewAppSwitcherReverse(x.Windows(), fynedesk.Instance().IconProvider())

			if ignoreSwitcher {
				ignoreSwitcher = false
			} else {
				switcherInstance = win
				win.Show()
			}
		} else {
			win := ui.NewAppSwitcher(x.Windows(), fynedesk.Instance().IconProvider())
			if ignoreSwitcher {
				ignoreSwitcher = false
			} else {
				switcherInstance = win
				win.Show()
			}
		}
	}()
}
