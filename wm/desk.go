// +build linux,!ci

package wm

import (
	"errors"
	"log"

	"fyne.io/fyne"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/fyne-io/desktop"
	"github.com/fyne-io/desktop/theme"
)

type x11WM struct {
	stack
	x      *xgbutil.XUtil
	root   fyne.Window
	rootID xproto.Window
	loaded bool
}

func (x *x11WM) Close() {
	log.Println("Disconnecting from X server")

	for _, child := range x.frames {
		child.(*frame).unFrame()
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
	root := conn.RootWin()
	eventMask := xproto.EventMaskPropertyChange |
		xproto.EventMaskFocusChange |
		xproto.EventMaskButtonPress |
		xproto.EventMaskButtonRelease |
		xproto.EventMaskKeyPress |
		xproto.EventMaskVisibilityChange |
		xproto.EventMaskStructureNotify |
		xproto.EventMaskSubstructureNotify |
		xproto.EventMaskSubstructureRedirect
	if err := xproto.ChangeWindowAttributesChecked(conn.Conn(), root, xproto.CwEventMask,
		[]uint32{uint32(eventMask)}).Check(); err != nil {
		conn.Conn().Close()

		return nil, errors.New("window manager detected, running embedded")
	}

	loadCursors(conn)
	go mgr.runLoop()

	listener := make(chan fyne.Settings)
	a.Settings().AddChangeListener(listener)
	go func() {
		for {
			<-listener
			for _, fr := range mgr.frames {
				fr.(*frame).ApplyTheme()
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
			case xproto.DestroyNotifyEvent:
				x.destroyWindow(ev.Window)
			case xproto.PropertyNotifyEvent:
				// TODO
			case xproto.ButtonPressEvent:
				for _, fr := range x.frames {
					if fr.(*frame).id == ev.Event {
						fr.(*frame).press(ev.EventX, ev.EventY)
					}
				}
			case xproto.ButtonReleaseEvent:
				for _, fr := range x.frames {
					if fr.(*frame).id == ev.Event {
						fr.(*frame).release(ev.EventX, ev.EventY)
					}
				}
			case xproto.MotionNotifyEvent:
				for _, fr := range x.frames {
					if fr.(*frame).id == ev.Event {
						if ev.State&xproto.ButtonMask1 != 0 {
							fr.(*frame).drag(ev.EventX, ev.EventY)
						} else {
							fr.(*frame).motion(ev.EventX, ev.EventY)
						}
					}
				}
			case xproto.KeyPressEvent:
				winList := x.Windows()
				winCount := len(winList)
				if winCount <= 1 {
					return
				}

				if ev.State&xproto.ModMaskShift != 0 {
					x.RaiseToTop(winList[winCount-1])
				} else {
					x.RaiseToTop(winList[1])
				}
			}
		}
	}
}

func (x *x11WM) configureWindow(win xproto.Window, ev xproto.ConfigureRequestEvent) {
	frame := x.frameForWin(win)
	width := ev.Width
	height := ev.Height

	if frame != nil && frame.Decorated() {
		err := xproto.ConfigureWindowChecked(x.x.Conn(), win, xproto.ConfigWindowX|xproto.ConfigWindowY|
			xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
			[]uint32{uint32(x.borderWidth()), uint32(x.borderWidth() + x.titleHeight()),
				uint32(width - 1), uint32(height - 1)}).Check()

		if err != nil {
			log.Println("ConfigureFrame Err", err)
		}
		return
	}

	name := windowName(x.x, win)
	isRoot := x.root != nil && name == x.root.Title()
	if isRoot {
		width = x.x.Screen().WidthInPixels
		height = x.x.Screen().HeightInPixels
	}

	err := xproto.ConfigureWindowChecked(x.x.Conn(), win, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(ev.X), uint32(ev.Y), uint32(width), uint32(height)}).Check()
	if err != nil {
		log.Println("ConfigureWindow Err", err)
	}

	if isRoot {
		if x.loaded {
			return
		}
		x.rootID = win
		x.loaded = true
	}
}

func (x *x11WM) showWindow(win xproto.Window) {
	framed := x.frameForWin(win) != nil
	name := windowName(x.x, win)

	if framed || name == x.root.Title() {
		err := xproto.MapWindowChecked(x.x.Conn(), win).Check()
		if err != nil {
			log.Println("ShowWindow Err", err)
		}

		x.bindKeys(win)
		go x.frameExisting()

		return
	}
	if x.rootID == 0 {
		return
	}

	x.setupWindow(win)
}

func (x *x11WM) hideWindow(win xproto.Window) {
	if x.frameForWin(win) != nil {
		fr := x.frameForWin(win)
		x.RemoveWindow(fr)
		xproto.UnmapWindow(x.x.Conn(), fr.(*frame).id)
	}
}

func (x *x11WM) setupWindow(win xproto.Window) {
	var frame *frame
	if !windowBorderless(x.x, win) {
		frame = newFrame(win, x)
	} else {
		frame = newFrameBorderless(win, x)
	}

	x.bindKeys(win)
	if x.root != nil && windowName(x.x, win) == x.root.Title() {
		return
	}

	x.AddWindow(frame)
	x.RaiseToTop(frame)
}

func (x *x11WM) destroyWindow(win xproto.Window) {
	xproto.ChangeSaveSet(x.x.Conn(), xproto.SetModeDelete, win)
}

func (x *x11WM) bindKeys(win xproto.Window) {
	xproto.GrabKey(x.x.Conn(), true, win, xproto.ModMask1, 23, xproto.GrabModeAsync, xproto.GrabModeAsync)
	xproto.GrabKey(x.x.Conn(), true, win, xproto.ModMaskShift|xproto.ModMask1, 23, xproto.GrabModeAsync, xproto.GrabModeAsync)
}

func (x *x11WM) frameExisting() {
	tree, err := xproto.QueryTree(x.x.Conn(), x.x.RootWin()).Reply()
	if err != nil {
		log.Println("QueryTree Err", err)
		return
	}

	for _, child := range tree.Children {
		name := windowName(x.x, child)
		if x.root != nil && name == x.root.Title() {
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

		x.setupWindow(child)
	}
}

func (x *x11WM) borderWidth() uint16 {
	scale := float32(1.0)
	if x.root != nil {
		scale = x.root.Canvas().Scale()
	}
	return uint16(float32(theme.BorderWidth) * scale)
}

func (x *x11WM) buttonWidth() uint16 {
	scale := float32(1.0)
	if x.root != nil {
		scale = x.root.Canvas().Scale()
	}
	return uint16(float32(theme.ButtonWidth) * scale)
}

func (x *x11WM) titleHeight() uint16 {
	scale := float32(1.0)
	if x.root != nil {
		scale = x.root.Canvas().Scale()
	}
	return uint16(float32(theme.TitleHeight) * scale)
}
