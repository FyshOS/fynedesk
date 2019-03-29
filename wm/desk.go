// +build linux,!ci

package wm

import (
	"errors"
	"log"

	"fyne.io/fyne"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/fyne-io/desktop"
)

const (
	borderWidth = 5
	titleHeight = 15
)

type x11WM struct {
	x      *xgbutil.XUtil
	root   fyne.Window
	rootID xproto.Window
	loaded bool

	frames map[xproto.Window]*frame
}

func (x *x11WM) Close() {
	log.Println("Disconnecting from X server")

	for _, child := range x.frames {
		child.unframe()
	}

	x.x.Conn().Close()
}

func (x *x11WM) SetRoot(win fyne.Window) {
	x.root = win
}

// NewX11WindowManager sets up a new X11 Window Manager to control a desktop in X11.
func NewX11WindowManager(a fyne.App) (desktop.WindowManager, error) {
	conn, err := xgbutil.NewConn()
	if err != nil {
		log.Println("Failed to connect to the XServer", err)
		return nil, err
	}

	mgr := &x11WM{x: conn}
	mgr.frames = make(map[xproto.Window]*frame)
	mgr.frameExisting()

	root := conn.RootWin()
	eventMask := xproto.EventMaskPropertyChange |
		xproto.EventMaskFocusChange |
		xproto.EventMaskButtonPress |
		xproto.EventMaskButtonRelease |
		xproto.EventMaskVisibilityChange |
		xproto.EventMaskStructureNotify |
		xproto.EventMaskSubstructureNotify |
		xproto.EventMaskSubstructureRedirect
	if err := xproto.ChangeWindowAttributesChecked(conn.Conn(), root, xproto.CwEventMask,
		[]uint32{uint32(eventMask)}).Check(); err != nil {
		conn.Conn().Close()

		return nil, errors.New("Window manager detected, running embedded")
	}

	go mgr.runLoop()

	listener := make(chan fyne.Settings)
	a.Settings().AddChangeListener(listener)
	go func() {
		for {
			<-listener
			for _, frame := range mgr.frames {
				frame.ApplyTheme()
			}
		}
	}()

	return mgr, nil
}

func (x *x11WM) runLoop() {
	conn := x.x.Conn()

	for {
		ev, err := conn.WaitForEvent()
		if err != nil {
			log.Println("X11 Error:", err)
			continue
		}

		if ev != nil {
			switch ev := ev.(type) {
			case xproto.MapRequestEvent:
				x.showWindow(ev.Window)
			case xproto.UnmapNotifyEvent:
				x.hideWindow(ev.Window)
			case xproto.ConfigureRequestEvent:
				x.configureWindow(ev.Window, ev)
			case xproto.PropertyNotifyEvent:
				if ev.Atom == xproto.AtomWmIconName {
					log.Println("Hints", ev.State)
				}
			}
		}
	}
}

func (x *x11WM) configureWindow(win xproto.Window, ev xproto.ConfigureRequestEvent) {
	framed := x.frames[win] != nil
	if framed {
		err := xproto.ConfigureWindowChecked(x.x.Conn(), x.frames[win].id, xproto.ConfigWindowX|xproto.ConfigWindowY|
			xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
			[]uint32{uint32(ev.X), uint32(ev.Y), uint32(ev.Width + borderWidth*2), uint32(ev.Height + borderWidth*2 + titleHeight)}).Check()

		if err != nil {
			log.Println("ConfigureFrame Err", err)
		}
	}
	err := xproto.ConfigureWindowChecked(x.x.Conn(), win, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(ev.X), uint32(ev.Y), uint32(ev.Width), uint32(ev.Height)}).Check()
	if err != nil {
		log.Println("ConfigureWindow Err", err)
	}

	prop, _ := xprop.GetProperty(x.x, win, "WM_NAME")
	if !framed && prop != nil && x.root != nil && string(prop.Value) == x.root.Title() {
		if x.loaded {
			return
		}
		x.rootID = win
		x.loaded = true

		for _, framed := range x.frames {
			err = xproto.ConfigureWindowChecked(x.x.Conn(), framed.id, xproto.ConfigWindowSibling|xproto.ConfigWindowStackMode,
				[]uint32{uint32(x.rootID), uint32(xproto.StackModeAbove)}).Check()
			if err != nil {
				log.Println("Restack Err", err)
			}
		}
	}
}

func (x *x11WM) showWindow(win xproto.Window) {
	framed := x.frames[win] != nil
	prop, err := xprop.GetProperty(x.x, win, "WM_NAME")
	if err != nil {
		log.Println("GetProperty Err", err)
	}
	if framed || string(prop.Value) == x.root.Title() {
		err := xproto.MapWindowChecked(x.x.Conn(), win).Check()
		if err != nil {
			log.Println("ShowWindow Err", err)
		}

		return
	}
	if x.rootID == 0 {
		return
	}

	x.frame(win)
}

func (x *x11WM) frame(win xproto.Window) {
	frame := newFrame(win, x)

	x.frames[win] = frame
}

func (x *x11WM) frameExisting() {
	tree, err := xproto.QueryTree(x.x.Conn(), x.x.RootWin()).Reply()
	if err != nil {
		log.Println("QueryTree Err", err)
		return
	}

	for _, child := range tree.Children {
		prop, _ := xprop.GetProperty(x.x, child, "WM_NAME")
		if prop == nil || x.root != nil && string(prop.Value) == x.root.Title() {
			continue
		}

		attrs, err := xproto.GetWindowAttributes(x.x.Conn(), child).Reply()
		if err != nil {
			log.Println("GetWindowAttributes Err", err)
			continue
		}
		if attrs.MapState == xproto.MapStateUnmapped {
			continue
		}

		x.frame(child)
	}
}

func (x *x11WM) hideWindow(win xproto.Window) {
	if x.frames[win] != nil {
		frame := x.frames[win]
		xproto.UnmapWindow(x.x.Conn(), frame.id)
	}
}
