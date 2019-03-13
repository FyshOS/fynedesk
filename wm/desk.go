// +build linux,!ci

package wm

import (
	"errors"
	"log"

	"fyne.io/fyne"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/fyne-io/desktop"
)

const (
	borderWidth = 5
	titleHeight = 15
)

type x11WM struct {
	x    *xgbutil.XUtil
	root fyne.Window

	frames map[xproto.Window]xproto.Window
}

func (x *x11WM) Close() {
	log.Println("Disconnecting from X server")
	// TODO unregister etc
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

	root := conn.RootWin()
	disp := xproto.Setup(conn.Conn()).Roots[0]

	eventMask := xproto.EventMaskPropertyChange |
		xproto.EventMaskFocusChange |
		xproto.EventMaskButtonPress |
		xproto.EventMaskButtonRelease |
		xproto.EventMaskVisibilityChange |
		xproto.EventMaskStructureNotify |
		xproto.EventMaskSubstructureNotify |
		xproto.EventMaskSubstructureRedirect

	win := xwindow.New(conn, root)
	if err := xproto.ChangeWindowAttributesChecked(conn.Conn(), root, xproto.CwEventMask,
		[]uint32{uint32(eventMask)}).Check(); err != nil {
		conn.Conn().Close()

		return nil, errors.New("Window manager detected, running embedded")
	}
	win.Destroy()

	log.Println("Connected to X server")
	log.Println("Default Root", disp.WidthInPixels, "x", disp.HeightInPixels)

	mgr := &x11WM{x: conn}
	mgr.frames = make(map[xproto.Window]xproto.Window)

	go mgr.runLoop()

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
				}
			}
		}
	}
}

func (x *x11WM) configureWindow(win xproto.Window, ev xproto.ConfigureRequestEvent) {
	if x.frames[win] != 0 {
		xproto.ConfigureWindow(x.x.Conn(), x.frames[win], xproto.ConfigWindowX|xproto.ConfigWindowY|
			xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
			[]uint32{uint32(ev.X), uint32(ev.Y), uint32(ev.Width + borderWidth*2), uint32(ev.Height + borderWidth*2 + titleHeight)})
	}
	xproto.ConfigureWindow(x.x.Conn(), win, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(ev.X), uint32(ev.Y), uint32(ev.Width), uint32(ev.Height)})
}

func (x *x11WM) showWindow(win xproto.Window) {
	prop, _ := xprop.GetProperty(x.x, win, "WM_NAME")
	if string(prop.Value) == x.root.Title() {
		xproto.MapWindow(x.x.Conn(), win)
		return
	}

	attrs, err := xproto.GetGeometry(x.x.Conn(), xproto.Drawable(win)).Reply()
	if err != nil {
		log.Println("Err", err)
		return
	}

	frame, err := xwindow.Generate(x.x)
	if err != nil {
		log.Println("Err", err)
		return
	}

	values := []uint32{0xffaa66, xproto.EventMaskStructureNotify |
		xproto.EventMaskSubstructureNotify | xproto.EventMaskSubstructureRedirect |
		xproto.EventMaskButtonPress | xproto.EventMaskButtonRelease |
		xproto.EventMaskFocusChange}
	err = xproto.CreateWindowChecked(x.x.Conn(), x.x.Screen().RootDepth, frame.Id, x.x.RootWin(),
		0, 0, attrs.Width+borderWidth*2, attrs.Height+borderWidth*2+titleHeight, 0, xproto.WindowClassInputOutput,
		x.x.Screen().RootVisual, xproto.CwBackPixel|xproto.CwEventMask, values).Check()
	if err != nil {
		log.Println("Err", err)
		return
	}

	x.frames[win] = frame.Id
	xproto.ReparentWindow(x.x.Conn(), win, frame.Id, borderWidth-1, borderWidth+titleHeight-1)
	frame.Map()
	xproto.MapWindow(x.x.Conn(), win)

	xproto.SetInputFocus(x.x.Conn(), 0, win, 0)
}

func (x *x11WM) hideWindow(win xproto.Window) {
	if x.frames[win] != 0 {
		frame := x.frames[win]
		xproto.UnmapWindow(x.x.Conn(), frame)
	}
}
