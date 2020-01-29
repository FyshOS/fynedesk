// +build linux

package wm

import (
	"fyne.io/desktop"
	"fyne.io/desktop/internal/ui"

	"github.com/BurntSushi/xgb/xproto"
)

// switcherInstance and the methods below manage the X to ui.Switcher bindings.
// This is needed due to the way that releasing Alt is only reported if we grab the whole keyboard.
// Therefore the UI cannot handle keyboard input - so we add it here instead.
var switcherInstance *ui.Switcher

func (x *x11WM) showOrSelectAppSwitcher(reverse bool) {
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
			switcherInstance = ui.ShowAppSwitcherReverse(x.Windows(), desktop.Instance().IconProvider())
		} else {
			switcherInstance = ui.ShowAppSwitcher(x.Windows(), desktop.Instance().IconProvider())
		}
	}()
}

func (x *x11WM) previousAppSwitcher() {
	if switcherInstance == nil {
		return
	}

	go switcherInstance.Previous()
}

func (x *x11WM) nextAppSwitcher() {
	if switcherInstance == nil {
		return
	}

	go switcherInstance.Next()
}

func (x *x11WM) cancelAppSwitcher() {
	if switcherInstance == nil {
		return
	}

	go switcherInstance.HideCancel()
	xproto.UngrabKeyboard(x.x.Conn(), xproto.TimeCurrentTime)
	switcherInstance = nil
}

func (x *x11WM) applyAppSwitcher() {
	if switcherInstance == nil {
		return
	}

	go switcherInstance.HideApply()
	xproto.UngrabKeyboard(x.x.Conn(), xproto.TimeCurrentTime)
	windowClientListStackingUpdate(x)
	switcherInstance = nil
}
