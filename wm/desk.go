// +build linux,!ci

package wm

import (
	"log"

	"fyne.io/fyne"
	"github.com/fyne-io/desktop"
	"github.com/pkg/errors"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xwindow"
)

type x11WM struct {
	x *xgbutil.XUtil
}

func (x *x11WM) Close() {
	log.Println("Disconnecting from X server")
	// TODO unregister etc
	x.x.Conn().Close()
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
	if err := win.Listen(eventMask); err != nil {
		conn.Conn().Close()

		return nil, errors.New("Window manager detected, running embedded")
	}
	win.Destroy()

	log.Println("Connected to X server")
	log.Println("Default Root", disp.WidthInPixels, "x", disp.HeightInPixels)

	mgr := &x11WM{x: conn}

	go func() {
		for {
			ev, err := conn.Conn().WaitForEvent()
			if err != nil {
				log.Println("X11 Error:", err)
				continue
			}

			if ev != nil {
				switch ev := ev.(type) {
				case xproto.MapRequestEvent:
					xproto.MapWindow(conn.Conn(), ev.Window)
					xproto.SetInputFocus(conn.Conn(), 0, ev.Window, 0)
				case xproto.ConfigureRequestEvent:
					xproto.ConfigureWindow(conn.Conn(), ev.Window, xproto.ConfigWindowX|xproto.ConfigWindowY|
						xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
						[]uint32{uint32(ev.X), uint32(ev.Y), uint32(ev.Width), uint32(ev.Height)})
				}
			}
		}
	}()

	return mgr, nil
}
